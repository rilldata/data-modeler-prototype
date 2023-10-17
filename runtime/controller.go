package runtime

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/dag"
	"github.com/rilldata/rill/runtime/pkg/schedule"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

// tracer to trace background reconile calls
var tracer = otel.Tracer("github.com/rilldata/rill/runtime/controller")

// errCyclicDependency is set as the error on resources that can't be reconciled due to a cyclic dependency
var errCyclicDependency = errors.New("cannot be reconciled due to cyclic dependency")

// errControllerNotRunning is returned from controller functions that require the controller to be running
var errControllerNotRunning = errors.New("controller is not running")

// Reconciler implements reconciliation logic for all resources of a specific kind.
// Reconcilers are managed and invoked by a Controller.
type Reconciler interface {
	Close(ctx context.Context) error
	AssignSpec(from, to *runtimev1.Resource) error
	AssignState(from, to *runtimev1.Resource) error
	ResetState(r *runtimev1.Resource) error
	Reconcile(ctx context.Context, n *runtimev1.ResourceName) ReconcileResult
}

// ReconcileResult propagates results from a reconciler invocation
type ReconcileResult struct {
	Err       error
	Retrigger time.Time
}

// ReconcilerInitializer is a function that initializes a new reconciler for a specific controller
type ReconcilerInitializer func(*Controller) Reconciler

// ReconcilerInitializers is a registry of reconciler initializers for different resource kinds.
// There can be only one reconciler per resource kind.
var ReconcilerInitializers = make(map[string]ReconcilerInitializer)

// RegisterReconciler registers a reconciler initializer for a specific resource kind
func RegisterReconcilerInitializer(resourceKind string, initializer ReconcilerInitializer) {
	if ReconcilerInitializers[resourceKind] != nil {
		panic(fmt.Errorf("reconciler already registered for resource kind %q", resourceKind))
	}
	ReconcilerInitializers[resourceKind] = initializer
}

// Controller manages the catalog for a single instance and runs reconcilers to migrate the catalog (and related resources in external databases) into the desired state.
// For information about how the controller schedules reconcilers, see `runtime/reconcilers/README.md`.
type Controller struct {
	Runtime     *Runtime
	InstanceID  string
	Logger      *slog.Logger
	Activity    activity.Client
	mu          sync.RWMutex
	running     atomic.Bool   // Indicates if the controller is running
	ready       chan struct{} // Closed when the controller transitions to running
	closed      chan struct{} // Closed when the controller is closed
	initErr     error         // error in initialising controller
	reconcilers map[string]Reconciler
	catalog     *catalogCache
	// subscribers tracks subscribers to catalog events.
	subscribers      map[int]SubscribeCallback
	nextSubscriberID int
	// idleWaits tracks goroutines waiting for the controller to become idle.
	idleWaits      map[int]idleWait
	nextIdleWaitID int
	// queue contains names waiting to be scheduled.
	// It's not a real queue because we usually schedule the whole queue on each call to processQueue.
	queue          map[string]*runtimev1.ResourceName
	queueUpdated   bool
	queueUpdatedCh chan struct{}
	// timeline tracks resources to be scheduled in the future.
	timeline *schedule.Schedule[string, *runtimev1.ResourceName]
	// invocations tracks currently running reconciler invocations.
	invocations map[string]*invocation
	// completed receives invocations that have finished running.
	completed chan *invocation
}

// NewController creates a new Controller
func NewController(rt *Runtime, instanceID string, logger *zap.Logger, ac activity.Client) (*Controller, error) {
	c := &Controller{
		Runtime:        rt,
		InstanceID:     instanceID,
		Activity:       ac,
		ready:          make(chan struct{}),
		closed:         make(chan struct{}),
		reconcilers:    make(map[string]Reconciler),
		subscribers:    make(map[int]SubscribeCallback),
		idleWaits:      make(map[int]idleWait),
		queue:          make(map[string]*runtimev1.ResourceName),
		queueUpdatedCh: make(chan struct{}, 1),
		timeline:       schedule.New[string, *runtimev1.ResourceName](nameStr),
		invocations:    make(map[string]*invocation),
		completed:      make(chan *invocation),
	}

	// Hacky way to customize the logger for local vs. hosted
	// TODO: Setup the logger to duplicate logs to a) the Zap logger, b) an in-memory buffer that exposes the logs over the API
	if !rt.AllowHostAccess() {
		logger = logger.With(zap.String("instance_id", instanceID))
		logger = logger.Named("console")
	}
	c.Logger = slog.New(zapslog.HandlerOptions{LoggerName: "console"}.New(logger.Core()))

	return c, nil
}

// Run starts and runs the controller's event loop. It returns when ctx is cancelled or an unrecoverable error occurs.
// The event loop schedules/invokes resource reconciliation and periodically flushes catalog changes to persistent storage.
// The implementation centers around these internal functions: enqueue, processQueue (uses markPending, trySchedule, invoke), and processCompletedInvocation.
// See their docstrings for further details.
func (c *Controller) Run(ctx context.Context) error {
	cc, err := newCatalogCache(ctx, c, c.InstanceID)
	if err != nil {
		c.initErr = err
		close(c.closed)
		return fmt.Errorf("failed to create catalog cache: %w", err)
	}
	c.catalog = cc

	if c.running.Swap(true) {
		panic("controller is already running")
	}

	// Check we are still the leader
	err = c.catalog.checkLeader(ctx)
	if err != nil {
		c.initErr = err
		close(c.closed)
		_ = c.catalog.close(ctx)
		return err
	}

	// Initially enqueue all resources
	c.mu.Lock()
	for _, rs := range c.catalog.resources {
		for _, r := range rs {
			c.enqueue(r.Meta.Name)
		}
	}
	c.mu.Unlock()

	// Ticker for periodically flushing catalog changes
	flushTicker := time.NewTicker(10 * time.Second)
	defer flushTicker.Stop()

	// Timer for scheduling resources added to c.timeline.
	// Call resetTimelineTimer whenever the timeline may have been changed (must hold mu).
	timelineTimer := time.NewTimer(time.Second)
	defer timelineTimer.Stop()
	timelineTimer.Stop() // We want it stopped initially
	nextTime := time.Time{}
	resetTimelineTimer := func() {
		_, t := c.timeline.Peek()
		if t == nextTime {
			return
		}
		nextTime = t

		timelineTimer.Stop()
		if t.IsZero() {
			return
		}

		d := time.Until(t)
		if d <= 0 {
			// must be positive
			d = time.Nanosecond
		}

		d += time.Second // Add a second to avoid rapid cancellations due to micro differences in schedule time
		timelineTimer.Reset(d)
	}

	// Open for business
	close(c.ready)

	// Run event loop
	var stop bool
	var loopErr error
	for !stop {
		select {
		case <-c.queueUpdatedCh: // There are resources we should schedule
			c.mu.Lock()
			err := c.processQueue()
			if err != nil {
				loopErr = err
				stop = true
			} else {
				resetTimelineTimer()
			}
			c.checkIdleWaits()
			c.mu.Unlock()
		case inv := <-c.completed: // A reconciler invocation has completed
			c.mu.Lock()
			err = c.processCompletedInvocation(inv)
			if err != nil {
				loopErr = err
				stop = true
			} else {
				resetTimelineTimer()
			}
			c.checkIdleWaits()
			c.mu.Unlock()
		case <-timelineTimer.C: // A previous reconciler invocation asked to be re-scheduled now
			c.mu.Lock()
			for c.timeline.Len() > 0 {
				n, t := c.timeline.Peek()
				if t.After(time.Now()) {
					break
				}
				c.timeline.Pop()
				c.enqueue(n)
			}
			resetTimelineTimer()
			c.mu.Unlock()
		case <-flushTicker.C: // It's time to flush the catalog to persistent storage
			c.mu.RLock()
			err = c.catalog.flush(ctx)
			c.mu.RUnlock()
			if err != nil {
				loopErr = err
				stop = true
			}
		case <-c.catalog.hasEventsCh: // The catalog has events to process
			// Need a write lock to call resetEvents.
			c.mu.Lock()
			events := c.catalog.events
			c.catalog.resetEvents()
			c.mu.Unlock()

			// Need a read lock to prevent c.subscribers from being modified while we're iterating over it.
			c.mu.RLock()
			for _, fn := range c.subscribers {
				for _, e := range events {
					fn(e.event, e.name, e.resource)
				}
			}
			c.mu.RUnlock()
		case <-ctx.Done(): // We've been asked to stop
			stop = true
			break
		}
	}

	// Cleanup time
	var closeErr error
	if loopErr != nil {
		closeErr = fmt.Errorf("controller event loop failed: %w", loopErr)
	}

	// Cancel all running invocations
	c.mu.RLock()
	for _, inv := range c.invocations {
		inv.cancel(false)
	}
	c.mu.RUnlock()

	// Allow 10 seconds for closing invocations and reconcilers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Need to consume all the cancelled invocation completions (otherwise, they will hang on sending to c.completed)
	for {
		c.mu.RLock()
		if len(c.invocations) == 0 {
			c.mu.RUnlock()
			break
		}
		c.mu.RUnlock()

		stop := false
		select {
		case inv := <-c.completed:
			c.mu.Lock()
			err = c.processCompletedInvocation(inv)
			if err != nil {
				c.Logger.Warn("failed to process completed invocation during shutdown", slog.Any("error", err))
			}
			c.mu.Unlock()
		case <-ctx.Done():
			err = fmt.Errorf("timed out waiting for reconcile to finish for resources: %v", maps.Keys(c.invocations))
			closeErr = errors.Join(closeErr, err)
			stop = true // can't use break inside a select
		}
		if stop {
			break
		}
	}

	// Close all reconcilers
	c.mu.Lock()
	for k, r := range c.reconcilers {
		err := r.Close(ctx)
		if err != nil {
			err = fmt.Errorf("failed to close reconciler for %q: %w", k, err)
			closeErr = errors.Join(closeErr, err)
		}
	}
	c.mu.Unlock()

	// Mark closed (no more catalog writes after this)
	c.running.Store(false)
	close(c.closed)

	// Ensure anything waiting for WaitUntilIdle is notified (not calling checkIdleWaits because the queue may not be empty when closing)
	c.mu.Lock()
	for _, iw := range c.idleWaits {
		close(iw.ch)
	}
	c.mu.Unlock()

	// Allow 10 seconds for flushing the catalog
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Close catalog cache (will call flush)
	c.mu.Lock()
	err = c.catalog.close(ctx)
	if err != nil {
		err = fmt.Errorf("failed to close catalog: %w", err)
		closeErr = errors.Join(closeErr, err)
	}
	c.mu.Unlock()

	if closeErr != nil {
		c.Logger.Error("controller closed with error", slog.Any("error", closeErr))
	}
	return closeErr
}

// WaitUntilReady returns when the controller is ready to process catalog operations
func (c *Controller) WaitUntilReady(ctx context.Context) error {
	select {
	case <-c.closed: // controller was closed even before it became ready
		return c.initErr
	case <-c.ready:
	case <-ctx.Done():
	}
	return ctx.Err()
}

// WaitUntilIdle returns when the controller is idle (i.e. no reconcilers are pending or running).
func (c *Controller) WaitUntilIdle(ctx context.Context, ignoreHidden bool) error {
	if err := c.checkRunning(); err != nil {
		return err
	}

	ch := make(chan struct{})

	c.mu.Lock()
	id := c.nextIdleWaitID
	c.nextIdleWaitID++
	c.idleWaits[id] = idleWait{ch: ch, ignoreHidden: ignoreHidden}
	c.checkIdleWaits() // we might be idle already, in which case this will immediately close the channel
	c.mu.Unlock()

	select {
	case <-ch:
		// No cleanup necessary - checkIdleWaits removes the wait from idleWaits
	case <-ctx.Done():
		// NOTE: Can't deadlock because ch is never sent to, only closed.
		c.mu.Lock()
		delete(c.idleWaits, id)
		c.mu.Unlock()
	}
	return ctx.Err()
}

// Get returns a resource by name.
// Soft-deleted resources (i.e. resources where DeletedOn != nil) are not returned.
func (c *Controller) Get(ctx context.Context, name *runtimev1.ResourceName, clone bool) (*runtimev1.Resource, error) {
	if err := c.checkRunning(); err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	c.lock(ctx, true)
	defer c.unlock(ctx, true)

	// We don't return soft-deleted resources, unless the lookup is from the reconciler itself (which may be executing the delete).
	withDeleted := c.isReconcilerForResource(ctx, name)

	return c.catalog.get(name, withDeleted, clone)
}

// List returns a list of resources of the specified kind.
// If kind is empty, all resources are returned.
// Soft-deleted resources (i.e. resources where DeletedOn != nil) are not returned.
func (c *Controller) List(ctx context.Context, kind string, clone bool) ([]*runtimev1.Resource, error) {
	if err := c.checkRunning(); err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	c.lock(ctx, true)
	defer c.unlock(ctx, true)
	return c.catalog.list(kind, false, clone)
}

// SubscribeCallback is the callback type passed to Subscribe.
type SubscribeCallback func(e runtimev1.ResourceEvent, n *runtimev1.ResourceName, r *runtimev1.Resource)

// Subscribe registers a callback that will receive resource update events.
// The same callback function will not be invoked concurrently.
// The callback function is invoked under a lock and must not call the controller.
func (c *Controller) Subscribe(ctx context.Context, fn SubscribeCallback) error {
	if err := c.checkRunning(); err != nil {
		return err
	}

	c.mu.Lock()
	id := c.nextSubscriberID
	c.nextSubscriberID++
	c.subscribers[id] = fn
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.subscribers, id)
		c.mu.Unlock()
	}()

	for {
		select {
		case <-c.closed:
			return fmt.Errorf("controller closed")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Create creates a resource and enqueues it for reconciliation.
// If a resource with the same name is currently being deleted, the deletion will be cancelled.
func (c *Controller) Create(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, hidden bool, r *runtimev1.Resource) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	// A deleted resource with the same name may exist and be running. If so, we first cancel it.
	requeued := false
	if inv, ok := c.invocations[nameStr(name)]; ok && !inv.deletedSelf {
		r, err := c.catalog.get(name, true, false)
		if err != nil {
			return fmt.Errorf("internal: got catalog error for reconciling resource: %w", err)
		}
		if r.Meta.DeletedOn == nil {
			// If a non-deleted resource exists with the same name, we should return an error instead of cancelling.
			return drivers.ErrResourceAlreadyExists
		}
		inv.cancel(true)
		requeued = true
	}

	err := c.catalog.create(name, refs, owner, paths, hidden, r)
	if err != nil {
		return err
	}

	if !requeued {
		c.enqueue(name)
	}
	return nil
}

// UpdateMeta updates a resource's meta fields and enqueues it for reconciliation.
// If called from outside the resource's reconciler and the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) UpdateMeta(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		c.cancelIfRunning(name, false)
		c.enqueue(name)
	}

	err := c.safeMutateRenamed(name)
	if err != nil {
		return err
	}

	err = c.catalog.updateMeta(name, refs, owner, paths)
	if err != nil {
		return err
	}

	// We updated refs, so it may have broken previous cyclic references
	ns := c.catalog.retryCyclicRefs()
	for _, n := range ns {
		c.enqueue(n)
	}

	return nil
}

// UpdateName renames a resource and updates annotations, and enqueues it for reconciliation.
// If called from outside the resource's reconciler and the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) UpdateName(ctx context.Context, name, newName, owner *runtimev1.ResourceName, paths []string) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		c.cancelIfRunning(name, false)
		c.enqueue(name)
	}

	// Check resource exists (otherwise, DAG lookup may panic)
	r, err := c.catalog.get(name, false, false)
	if err != nil {
		return err
	}

	// All resources pointing to the old name need to be reconciled (they'll pointing to a void resource after this)
	if !c.catalog.isCyclic(name) {
		ns := c.catalog.dag.Children(name)
		for _, n := range ns {
			c.enqueue(n)
		}
	}

	err = c.safeRename(name, newName)
	if err != nil {
		return err
	}
	c.enqueue(newName)

	err = c.catalog.updateMeta(newName, r.Meta.Refs, owner, paths)
	if err != nil {
		return err
	}

	// We updated a name, so it may have broken previous cyclic references
	ns := c.catalog.retryCyclicRefs()
	for _, n := range ns {
		c.enqueue(n)
	}

	return nil
}

// UpdateSpec updates a resource's spec and enqueues it for reconciliation.
// If called from outside the resource's reconciler and the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) UpdateSpec(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		c.cancelIfRunning(name, false)
		c.enqueue(name)
	}

	err := c.safeMutateRenamed(name)
	if err != nil {
		return err
	}

	err = c.catalog.updateSpec(name, r)
	if err != nil {
		return err
	}

	return nil
}

// UpdateState updates a resource's state.
// It can only be called from within the resource's reconciler.
// NOTE: Calls to UpdateState succeed even if ctx is cancelled. This enables cancelled reconcilers to update state before finishing.
func (c *Controller) UpdateState(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	// Must not check ctx.Err(). See NOTE above.
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		return fmt.Errorf("can't update resource state from outside of reconciler")
	}

	err := c.catalog.updateState(name, r)
	if err != nil {
		return err
	}

	return nil
}

// UpdateError updates a resource's error.
// Unlike UpdateMeta and UpdateSpec, it does not cancel or enqueue reconciliation for the resource.
func (c *Controller) UpdateError(ctx context.Context, name *runtimev1.ResourceName, reconcileErr error) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	err := c.catalog.updateError(name, reconcileErr)
	if err != nil {
		return err
	}

	return nil
}

// Delete soft-deletes a resource and enqueues it for reconciliation (with DeletedOn != nil).
// Once the deleting reconciliation has been completed, the resource will be hard deleted.
// If Delete is called from the resource's own reconciler, the resource will be hard deleted immediately (and the calling reconcile's ctx will be canceled immediately).
func (c *Controller) Delete(ctx context.Context, name *runtimev1.ResourceName) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	c.cancelIfRunning(name, false)

	// Check resource exists (otherwise, DAG lookup may panic)
	_, err := c.catalog.get(name, false, false)
	if err != nil {
		return err
	}

	// All resources pointing to deleted resource need to be reconciled (they'll pointing to a void resource after this)
	if !c.catalog.isCyclic(name) {
		ns := c.catalog.dag.Children(name)
		for _, n := range ns {
			c.enqueue(n)
		}
	}

	if c.isReconcilerForResource(ctx, name) {
		inv := invocationFromContext(ctx)
		inv.deletedSelf = true
		err := c.catalog.delete(name)
		if err != nil {
			return err
		}
	} else {
		err := c.catalog.clearRenamedFrom(name) // Avoid resource being marked both deleted and renamed
		if err != nil {
			return err
		}

		err = c.catalog.updateDeleted(name)
		if err != nil {
			return err
		}

		c.enqueue(name)
	}

	// We removed a name, so it may have broken previous cyclic references
	ns := c.catalog.retryCyclicRefs()
	for _, n := range ns {
		c.enqueue(n)
	}

	return nil
}

// Flush forces a flush of the controller's catalog changes to persistent storage.
func (c *Controller) Flush(ctx context.Context) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	return c.catalog.flush(ctx)
}

// Reconcile enqueues a resource for reconciliation.
// If the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) Reconcile(ctx context.Context, name *runtimev1.ResourceName) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)
	c.enqueue(name)
	return nil
}

// Cancel cancels the current invocation of a resource's reconciler (if it's running).
// It does not re-enqueue the resource for reconciliation.
func (c *Controller) Cancel(ctx context.Context, name *runtimev1.ResourceName) error {
	if err := c.checkRunning(); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	c.lock(ctx, false)
	defer c.unlock(ctx, false)
	c.cancelIfRunning(name, false)
	return nil
}

// AcquireOLAP gets a handle for a connector in the controller's instance.
func (c *Controller) AcquireConn(ctx context.Context, connector string) (drivers.Handle, func(), error) {
	return c.Runtime.AcquireHandle(ctx, c.InstanceID, connector)
}

// AcquireOLAP gets an OLAP handle for a connector in the controller's instance.
func (c *Controller) AcquireOLAP(ctx context.Context, connector string) (drivers.OLAPStore, func(), error) {
	conn, release, err := c.AcquireConn(ctx, connector)
	if err != nil {
		return nil, nil, err
	}

	olap, ok := conn.AsOLAP(c.InstanceID)
	if !ok {
		release()
		return nil, nil, fmt.Errorf("connector %q is not an OLAP", connector)
	}

	return olap, release, nil
}

// Lock locks the controller's catalog and delays scheduling of new reconciliations until the lock is released.
// It can only be called from within a reconciler invocation.
// While the lock is held, resources can only be edited by a caller using the ctx passed to Lock.
func (c *Controller) Lock(ctx context.Context) {
	inv := invocationFromContext(ctx)
	if inv == nil {
		panic("Lock called outside of a reconciler invocation")
	}
	if inv.holdsLock {
		panic("Lock called by invocation that already holds the lock")
	}
	inv.holdsLock = true
	c.mu.Lock()
}

// Unlock releases the lock acquired by Lock.
func (c *Controller) Unlock(ctx context.Context) {
	inv := invocationFromContext(ctx)
	if inv == nil {
		panic("Unlock called outside of a reconciler invocation")
	}
	if !inv.holdsLock {
		panic("Unlock called by invocation that does not hold the lock")
	}
	inv.holdsLock = false
	c.mu.Unlock()
}

// reconciler gets or lazily initializes a reconciler.
// reconciler is not thread-safe and must be called while c.mu is held.
func (c *Controller) reconciler(resourceKind string) Reconciler {
	reconciler := c.reconcilers[resourceKind]
	if reconciler != nil {
		return reconciler
	}

	initializer := ReconcilerInitializers[resourceKind]
	if initializer == nil {
		panic(fmt.Errorf("no reconciler registered for resource kind %q", resourceKind))
	}

	reconciler = initializer(c)
	c.reconcilers[resourceKind] = reconciler

	return reconciler
}

// checkRunning panics if called when the Controller is not running.
func (c *Controller) checkRunning() error {
	if !c.running.Load() {
		return errControllerNotRunning
	}
	return nil
}

// idleWait represents a caller waiting for the controller to become idle.
// If ignoreHidden is true, the controller will be considered idle if all running invocations are for hidden resources.
type idleWait struct {
	ch           chan struct{}
	ignoreHidden bool
}

// checkIdleWaits checks registered idleWaits and removes any that can be satisfied.
// It must be called with c.mu held.
func (c *Controller) checkIdleWaits() {
	if len(c.idleWaits) == 0 {
		return
	}

	// Generally, we're idle if: len(c.queue) == 0 && len(c.invocations) == 0
	// But we need to do some other extra checks to handle ignoreHidden.

	// The queue is processed rapidly, so we don't check ignoreHidden against it.
	if len(c.queue) != 0 {
		return // Not idle
	}

	// Look for non-hidden invocations
	found := false
	for _, inv := range c.invocations {
		if inv.isHidden {
			continue
		}
		found = true
		break
	}
	if found {
		return // Not idle
	}
	// We now know that all invocations (if any) are hidden.

	// Check individual waits
	for k, iw := range c.idleWaits {
		if !iw.ignoreHidden && len(c.invocations) != 0 {
			continue
		}

		delete(c.idleWaits, k)
		close(iw.ch)
	}
}

// lock locks the controller's mutex, unless ctx belongs to a reconciler invocation that already holds the lock (by having called c.Lock).
func (c *Controller) lock(ctx context.Context, read bool) {
	inv := invocationFromContext(ctx)
	if inv != nil && inv.holdsLock {
		return
	}
	if read {
		c.mu.RLock()
	} else {
		c.mu.Lock()
	}
}

// unlock unlocks the controller's mutex, unless ctx belongs to a reconciler invocation that holds the lock (by having called c.Lock).
func (c *Controller) unlock(ctx context.Context, read bool) {
	inv := invocationFromContext(ctx)
	if inv != nil && inv.holdsLock {
		return
	}
	if read {
		c.mu.RUnlock()
	} else {
		c.mu.Unlock()
	}
}

// isReconcilerForResource returns true if ctx belongs to a reconciler invocation for the given resource.
func (c *Controller) isReconcilerForResource(ctx context.Context, n *runtimev1.ResourceName) bool {
	inv := invocationFromContext(ctx)
	if inv == nil {
		return false
	}
	return inv.name.Kind == n.Kind && strings.EqualFold(inv.name.Name, n.Name) // NOTE: More efficient, but equivalent to: nameStr(inv.name) == nameStr(n)
}

// safeMutateRenamed makes it safe to mutate a resource that's currently being renamed by changing the rename to a delete+create.
// It does nothing if the resource is not currently being renamed (RenamedFrom == nil).
// It must be called while c.mu is held.
func (c *Controller) safeMutateRenamed(n *runtimev1.ResourceName) error {
	r, err := c.catalog.get(n, true, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return nil
		}
		return err
	}

	renamedFrom := r.Meta.RenamedFrom
	if renamedFrom == nil {
		return nil
	}

	err = c.catalog.clearRenamedFrom(n)
	if err != nil {
		return err
	}

	_, err = c.catalog.get(renamedFrom, true, false)
	if err == nil {
		// Either a new resource with the name of the old one has been created in the mean time,
		// or the rename just changed the casing of the name.
		// In either case, no delete is necessary (reconciler will bring to desired state).
		return nil
	}

	// Create a new resource with the old name, so we can delete it separately.
	err = c.catalog.create(renamedFrom, r.Meta.Refs, r.Meta.Owner, r.Meta.FilePaths, r.Meta.Hidden, r)
	if err != nil {
		return err
	}

	err = c.catalog.updateDeleted(renamedFrom)
	if err != nil {
		return err
	}

	c.enqueue(renamedFrom)
	return nil
}

// safeRename safely renames a resource, handling the case where multiple resources are renamed at the same time with collisions between old and new names.
// For example, imagine there are resources A and B, and then A is renamed to B and B is renamed to C simultaneously.
// safeRename resolves collisions by changing some renames to deletes+creates, which works because processQueue ensures deletes are run before creates and renames.
// It must be called while c.mu is held.
func (c *Controller) safeRename(from, to *runtimev1.ResourceName) error {
	// Just to be safe.
	// NOTE: Not a case insensitive comparison, since we actually want to rename in cases where the casing changed.
	if proto.Equal(from, to) {
		return nil
	}

	// There's a collision if to matches RenamedFrom of another resource.
	collision := false
	for _, n := range c.catalog.renamed {
		r, err := c.catalog.get(n, true, false)
		if err != nil {
			return fmt.Errorf("internal: failed to get renamed resource %v: %w", n, err)
		}
		if nameStr(to) == nameStr(r.Meta.RenamedFrom) {
			collision = true
			break
		}
	}

	// No collision, do a normal rename
	if !collision {
		return c.catalog.rename(from, to)
	}

	// There's a collision.

	// Handle the case where a resource was renamed from e.g. Aa to AA, and then while reconciling, is again renamed from AA to aA.
	// In this case, we still do a normal rename and rely on the reconciler to sort it out.
	if nameStr(from) == nameStr(to) {
		return c.catalog.rename(from, to)
	}

	// Do a create+delete instead of a rename.
	// This is safe because processQueue ensures deletes are run before creates.
	// NOTE: Doing the create first, since creation might fail if the name is taken, whereas the delete is almost certain to succeed.
	r, err := c.catalog.get(from, true, false)
	if err != nil {
		return err
	}

	err = c.catalog.create(to, r.Meta.Refs, r.Meta.Owner, r.Meta.FilePaths, r.Meta.Hidden, r)
	if err != nil {
		return err
	}

	err = c.catalog.updateDeleted(from)
	if err != nil {
		return err
	}

	c.enqueue(from)
	// The caller of safeRename will enqueue the new name

	return nil
}

// enqueue marks a resource to be scheduled in the next iteration of the event loop.
// It does so by adding it to c.queue, which will be processed by processQueue().
// It must be called while c.mu is held.
func (c *Controller) enqueue(name *runtimev1.ResourceName) {
	c.queue[nameStr(name)] = name
	c.setQueueUpdated()
}

// setQueueUpdated notifies the event loop that the queue has been updated and needs to be processed.
// It must be called while c.mu is held.
func (c *Controller) setQueueUpdated() {
	if !c.queueUpdated {
		c.queueUpdated = true
		c.queueUpdatedCh <- struct{}{}
	}
}

// processQueue calls attempts to schedule the resources in c.queue. It is invoked in each iteration of the event loop when there are resources in the queue.
// The reason we have the queue and process it from the event loop (instead of marking pending and scheduling directly from enqueue()) is to enable batch scheduling during initialization and when Lock/Unlock is used.
// Batching makes it easier to ensure parents are scheduled before children when both are enqueued at the same time.
//
// It must be called while c.mu is held.
func (c *Controller) processQueue() error {
	// Mark-sweep like approach - first mark all impacted resources (including descendents) pending, then schedule the ones that have no pending parents.

	// Phase 1: Mark items pending and trim queue when possible.
	for s, n := range c.queue {
		skip, err := c.markPending(n)
		if err != nil {
			return err
		}
		if skip {
			delete(c.queue, s)
		}
	}

	// Phase 2: Ensure scheduling
	for s, n := range c.queue {
		ok, err := c.trySchedule(n)
		if err != nil {
			return err
		}
		if ok {
			delete(c.queue, s)
		}
	}

	// Reset queueUpdated
	c.queueUpdated = false
	return nil
}

// markPending marks a resource and its descendents as pending.
// It also clears errors on every resource marked pending - it would be confusing to show an old error after a change has been made that may fix it.
// It returns true if it already now knows that the resource can't be scheduled and will be re-triggered later (e.g. by being added to a waitlist).
// It must be called while c.mu is held.
func (c *Controller) markPending(n *runtimev1.ResourceName) (skip bool, err error) {
	// Remove from timeline (if present)
	c.timeline.Remove(n)

	// Get resource
	r, err := c.catalog.get(n, true, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return true, nil
		}
		return false, err
	}

	// If currently running, cancel and reschedule when cancellation is done.
	// NOTE: We know children are already marked PENDING.
	inv, ok := c.invocations[nameStr(n)]
	if ok {
		inv.cancel(true)
		return true, nil
	}

	// Not running - clear error and mark pending
	err = c.catalog.updateError(n, nil)
	if err != nil {
		return false, err
	}
	err = c.catalog.updateStatus(n, runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING, time.Time{})
	if err != nil {
		return false, err
	}

	// Skipping cycle and descendent checks if it's a resource deletion (because deleted resources are not tracked in the DAG)
	if r.Meta.DeletedOn != nil {
		return false, nil
	}

	// If resource is cyclic, set error and skip it
	if c.catalog.isCyclic(n) {
		err = c.catalog.updateError(n, errCyclicDependency)
		if err != nil {
			return false, err
		}
		err = c.catalog.updateStatus(n, runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE, time.Time{})
		if err != nil {
			return false, err
		}
		if !r.Meta.Hidden {
			logArgs := []any{slog.String("name", n.Name), slog.String("kind", unqualifiedKind(n.Kind)), slog.Any("error", errCyclicDependency)}
			c.Logger.Error("Skipping resource", logArgs...)
		}
		return true, nil
	}

	// Ensure all descendents get marked pending and cancel any running descendents.
	// NOTE: DAG access is safe because we have already checked for cyclic.
	descendentRunning := false
	err = c.catalog.dag.Visit(n, func(ds string, dn *runtimev1.ResourceName) error {
		dr, err := c.catalog.get(dn, true, false)
		if err != nil {
			return fmt.Errorf("error getting dag node %q: %w", ds, err)
		}
		switch dr.Meta.ReconcileStatus {
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
			// Clear error and mark it pending
			err = c.catalog.updateError(n, nil)
			if err != nil {
				return fmt.Errorf("error updating dag node %q: %w", ds, err)
			}
			err = c.catalog.updateStatus(dn, runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING, time.Time{})
			if err != nil {
				return fmt.Errorf("error updating dag node %q: %w", ds, err)
			}
			return nil
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
			// If it's pending, we know all its descendents are also pending.
			// We still need to traverse it to know if any of its descendents are running, but can skip that if we already know a descendent is running (minor optimization).
			if descendentRunning {
				return dag.ErrSkip
			}
			return nil
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING:
			// Cancel it
			inv, ok := c.invocations[nameStr(dn)]
			if !ok {
				return fmt.Errorf("internal: no invocation found for resource %q with status=running", ds)
			}
			inv.cancel(false)                        // False means it will go IDLE, but with n in the waitlist it will be marked PENDING again in the next iteration.
			inv.addToWaitlist(n, r.Meta.SpecVersion) // Ensures n will get revisited when the cancellation is done.
			descendentRunning = true
			return dag.ErrSkip // No need to traverse its children - we know they're all pending.
		default:
			panic(fmt.Errorf("internal: unexpected status %v", dr.Meta.ReconcileStatus))
		}
	})
	if err != nil {
		return false, err
	}

	// If a descendent is running, remove from queue (will be re-added when descendent's reconcile returns)
	if descendentRunning {
		return true, nil
	}

	// Proceed to trySchedule
	return false, nil
}

// trySchedule schedules a resource for reconciliation. It should only be called from processQueue().
// It must be called while c.mu is held.
//
// It returns true if the resource was invoked OR if it knows it will eventually be reconciled through one of the enqueueing rules implemented in processCompletedInvocation (waitlist or enqueuing of children).
// It returns false if the resource can't be scheduled right now and should be retried later (kept in the queue).
//
// The implementation relies on the key invariant that all resources awaiting to be reconciled have status=pending, *including descendents of a resource with status=pending*.
// This is ensured through the assignment of status=pending in markPending.
func (c *Controller) trySchedule(n *runtimev1.ResourceName) (success bool, err error) {
	r, err := c.catalog.get(n, true, false)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return true, nil
		}
		return false, err
	}

	// Return true if any parents are pending or running.
	// NOTE 1: Only getting parents if it's not a deletion (deleted resources are not tracked in the DAG).
	// NOTE 2: DAG access is safe because markPending ensures we never trySchedule a cyclic resource.
	var parents []*runtimev1.ResourceName
	if r.Meta.DeletedOn == nil {
		parents = c.catalog.dag.Parents(n, true)
	}
	for _, pn := range parents {
		p, err := c.catalog.get(pn, false, false)
		if err != nil {
			return false, fmt.Errorf("internal: error getting present parent %q: %w", nameStr(pn), err)
		}
		if p.Meta.ReconcileStatus != runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE {
			// When the parent has completed running, processCompletedInvocation will enqueue its children and we'll run again.
			return true, nil
		}
	}

	// We want deletes to run before renames or regular reconciles.
	// And we want renames to run before regular reconciles.
	// Return false if there are deleted or renamed resources, and this isn't one of them.
	// (The resource will be kept in the queue and retried next time processQueue runs.)
	if r.Meta.DeletedOn == nil {
		if len(c.catalog.deleted) != 0 {
			return false, nil
		}

		if r.Meta.RenamedFrom == nil && len(c.catalog.renamed) != 0 {
			return false, nil
		}
	}

	// Invoke
	err = c.invoke(r)
	if err != nil {
		return false, err
	}
	return true, nil
}

// invoke starts a goroutine for reconciling the resource and tracks the invocation in c.invocations.
// It must be called while c.mu is held.
func (c *Controller) invoke(r *runtimev1.Resource) error {
	// Set status to running
	n := r.Meta.Name
	err := c.catalog.updateStatus(n, runtimev1.ReconcileStatus_RECONCILE_STATUS_RUNNING, time.Time{})
	if err != nil {
		return fmt.Errorf("error updating dag node %q: %w", nameStr(n), err)
	}

	// Track invocation
	ctx, cancel := context.WithCancel(context.Background())
	inv := &invocation{
		name:      n,
		isHidden:  r.Meta.Hidden,
		isDelete:  r.Meta.DeletedOn != nil,
		isRename:  r.Meta.RenamedFrom != nil,
		startedOn: time.Now(),
		cancelFn:  cancel,
	}
	c.invocations[nameStr(n)] = inv

	// Log invocation
	if !inv.isHidden {
		logArgs := []any{slog.String("name", n.Name), slog.String("kind", unqualifiedKind(n.Kind))}
		if inv.isDelete {
			logArgs = append(logArgs, slog.Bool("deleted", inv.isDelete))
		}
		if inv.isRename {
			logArgs = append(logArgs, slog.String("renamed_from", r.Meta.RenamedFrom.Name))
		}
		c.Logger.Info("Reconciling resource", logArgs...)
	}

	// Start reconcile in background
	ctx = contextWithInvocation(ctx, inv)
	reconciler := c.reconciler(n.Kind) // fetched outside of goroutine to keep access under mutex
	go func() {
		defer func() {
			// Catch panics and set as error
			if r := recover(); r != nil {
				stack := make([]byte, 64<<10)
				stack = stack[:runtime.Stack(stack, false)]
				c.Logger.Error("panic in reconciler", slog.String("name", n.Name), slog.String("kind", n.Kind), slog.Any("error", r), slog.String("stack", string(stack)))

				inv.result = ReconcileResult{Err: fmt.Errorf("panic: %v", r)}
				if inv.holdsLock {
					c.Unlock(ctx)
				}
			}
			// Ensure ctx cancel is called (just for cleanup)
			cancel()
			// Send invocation to event loop for post-processing
			c.completed <- inv
		}()
		// Start tracing span
		tracerAttrs := []attribute.KeyValue{
			attribute.String("instance_id", c.InstanceID),
			attribute.String("name", n.Name),
			attribute.String("kind", unqualifiedKind(n.Kind)),
		}
		if inv.isDelete {
			tracerAttrs = append(tracerAttrs, attribute.Bool("deleted", inv.isDelete))
		}
		if inv.isRename {
			tracerAttrs = append(tracerAttrs, attribute.String("renamed_from", r.Meta.RenamedFrom.Name))
		}
		ctx, span := tracer.Start(ctx, "reconcile", trace.WithAttributes(tracerAttrs...))
		defer span.End()

		// Invoke reconciler
		inv.result = reconciler.Reconcile(ctx, n)
	}()

	return nil
}

// processCompletedInvocation does post-processing after a reconciler invocation completes.
// It must be called while c.mu is held.
//
// It updates the catalog with the invocation result and it calls enqueue() for any resources that are unblocked by its completion.
// It calls enqueue() based on the following rules:
// - for all the resources on its waitlist
// - and, for itself if inv.reschedule is true
// - and, for its children in the DAG if inv.reschedule is false
func (c *Controller) processCompletedInvocation(inv *invocation) error {
	// Cleanup - must remove it from c.invocations before any error conditions can occur (otherwise, closing the event loop can hang)
	delete(c.invocations, nameStr(inv.name))

	// Log result
	logArgs := []any{
		slog.String("name", inv.name.Name),
		slog.String("kind", unqualifiedKind(inv.name.Kind)),
	}
	elapsed := time.Since(inv.startedOn).Round(time.Millisecond)
	if elapsed > 0 {
		logArgs = append(logArgs, slog.Duration("elapsed", elapsed))
	}
	if !inv.result.Retrigger.IsZero() {
		logArgs = append(logArgs, slog.String("retrigger_on", inv.result.Retrigger.Format(time.RFC3339)))
	}
	if inv.cancelled {
		logArgs = append(logArgs, slog.Bool("cancelled", inv.cancelled))
	}
	errorLevel := false
	if inv.result.Err != nil && !errors.Is(inv.result.Err, context.Canceled) {
		logArgs = append(logArgs, slog.Any("error", inv.result.Err))
		errorLevel = true
	}
	if errorLevel {
		c.Logger.Error("Reconcile failed", logArgs...)
	} else if !inv.isHidden {
		c.Logger.Info("Reconciled resource", logArgs...)
	}

	r, err := c.catalog.get(inv.name, true, false)
	if err != nil {
		// Self-deletes are immediately hard deletes. So only return the error if it's not a self-delete.
		if !(inv.deletedSelf && errors.Is(err, drivers.ErrResourceNotFound)) {
			return err
		}
	}
	// NOTE: Due to self-deletes, r may be nil!

	if inv.isDelete {
		// Extra checks in case item was re-created during deletion, or deleted during a normal reconciling (in which case this is just a cancellation of the normal reconcile, not the result of deletion)
		if r != nil && r.Meta.DeletedOn != nil && !inv.cancelled {
			if inv.result.Err != nil {
				c.Logger.Error("got error while deleting resource", slog.String("name", nameStr(r.Meta.Name)), slog.Any("error", inv.result.Err))
			}

			err = c.catalog.delete(r.Meta.Name)
			if err != nil {
				return err
			}

			r = nil
		}

		// Trigger processQueue - there may be items in the queue waiting for all deletes to finish
		if len(c.catalog.deleted) == 0 {
			c.setQueueUpdated()
		}
	}
	// NOTE: r will be nil after here if it was deleted. Continuing in case there's a waitlist waiting for the deletion.

	if inv.isRename {
		// Extra checks in case item was cancelled during renaming
		if r != nil && r.Meta.RenamedFrom != nil && !inv.cancelled {
			err = c.catalog.clearRenamedFrom(r.Meta.Name)
			if err != nil {
				return err
			}
		}

		// Trigger processQueue - there may be items in the queue waiting for all renames to finish
		if len(c.catalog.renamed) == 0 {
			c.setQueueUpdated()
		}
	}

	// Track retrigger time
	if r != nil && !inv.result.Retrigger.IsZero() {
		if inv.result.Retrigger.After(time.Now()) {
			c.timeline.Set(r.Meta.Name, inv.result.Retrigger)
		} else {
			// If retrigger requested before now, we'll just reschedule directly
			inv.reschedule = true
			inv.result.Retrigger = time.Time{}
		}
	}

	// Update status and error (unless r was deleted, in which case r == nil)
	if r != nil {
		err = c.catalog.updateStatus(r.Meta.Name, runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE, inv.result.Retrigger)
		if err != nil {
			return err
		}
		err = c.catalog.updateError(r.Meta.Name, inv.result.Err)
		if err != nil {
			return err
		}
	}

	// Let the dominos fall
	if r != nil && inv.reschedule {
		c.enqueue(inv.name)
	}

	// Enqueue items from waitlist that haven't been updated (and hence re-triggered in the meantime).
	for _, e := range inv.waitlist {
		wr, err := c.catalog.get(e.name, true, false)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return err
		}
		if wr.Meta.SpecVersion == e.specVersion {
			c.enqueue(e.name)
		}
	}

	// Re-enqueue children if:
	if !inv.reschedule && // Not rescheduling  (since then the children would be blocked anyway)
		r != nil && // Not a hard delete (children were already enqueued when the soft delete happened)
		r.Meta.DeletedOn == nil && // Not a soft delete (children were already enqueued when c.Delete(...) was called)
		!c.catalog.isCyclic(inv.name) && // Hasn't become cyclic (since DAG access is not safe for cyclic names)
		true {
		for _, rn := range c.catalog.dag.Children(inv.name) {
			c.enqueue(rn)
		}
	}

	return nil
}

// cancelIfRunning cancels a running invocation for the resource.
// It does nothing if no invocation is running for the resource.
// It must be called while c.mu is held.
func (c *Controller) cancelIfRunning(n *runtimev1.ResourceName, reschedule bool) {
	inv, ok := c.invocations[nameStr(n)]
	if ok {
		inv.cancel(reschedule)
	}
}

// invocation represents a running reconciler invocation for a specific resource.
type invocation struct {
	name        *runtimev1.ResourceName
	isHidden    bool
	isDelete    bool
	isRename    bool
	startedOn   time.Time
	cancelFn    context.CancelFunc
	cancelled   bool
	reschedule  bool
	holdsLock   bool
	deletedSelf bool
	waitlist    map[string]waitlistEntry
	result      ReconcileResult
}

// waitlistEntry represents an entry in an invocation's waitlist.
type waitlistEntry struct {
	name        *runtimev1.ResourceName
	specVersion int64
}

// cancel cancels the invocation.
// It can be called multiple times with different reschedule values, and will be rescheduled if any of the calls ask for it.
// It's not thread-safe (must be called while the controller's mutex is held).
func (i *invocation) cancel(reschedule bool) {
	if !i.cancelled {
		i.cancelled = true
		i.cancelFn()
	}
	i.reschedule = i.reschedule || reschedule
}

// addToWaitlist adds a resource name to the invocation's waitlist.
// Resources on the waitlist will be scheduled after the invocation completes.
// It's not thread safe (must be called while the controller's mutex is held).
func (i *invocation) addToWaitlist(n *runtimev1.ResourceName, specVersion int64) {
	if i.waitlist == nil {
		i.waitlist = make(map[string]waitlistEntry)
	}
	i.waitlist[nameStr(n)] = waitlistEntry{
		name:        n,
		specVersion: specVersion,
	}
}

// invocationCtxKey is used for storing an invocation in a context.
type invocationCtxKey struct{}

// contextWithInvocation returns a wrapped context that contains an invocation.
func contextWithInvocation(ctx context.Context, inv *invocation) context.Context {
	return context.WithValue(ctx, invocationCtxKey{}, inv)
}

// invocationFromContext retrieves an invocation from a context.
// If no invocation is in the context, it returns nil.
func invocationFromContext(ctx context.Context) *invocation {
	inv := ctx.Value(invocationCtxKey{})
	if inv != nil {
		return inv.(*invocation)
	}
	return nil
}

// unqualifiedKind removes the protobuf package name from a kind.
// E.g. "rill.runtime.v1.Source" -> "Source".
func unqualifiedKind(k string) string {
	idx := strings.LastIndex(k, ".")
	if idx >= 0 {
		k = k[idx+1:]
	}
	// TEMP: Trim the "V2" suffix. TODO: Remove when dropping the suffixes.
	k = strings.TrimSuffix(k, "V2")
	return k
}
