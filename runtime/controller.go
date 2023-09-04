package runtime

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/dag2"
	"github.com/rilldata/rill/runtime/pkg/schedule"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"golang.org/x/exp/slog"
	"google.golang.org/protobuf/proto"
)

// Built-in resource kinds
const (
	ResourceKindProjectParser  string = "rill.runtime.v1.ProjectParser"
	ResourceKindSource         string = "rill.runtime.v1.SourceV2"
	ResourceKindModel          string = "rill.runtime.v1.ModelV2"
	ResourceKindMetricsView    string = "rill.runtime.v1.MetricsViewV2"
	ResourceKindMigration      string = "rill.runtime.v1.Migration"
	ResourceKindPullTrigger    string = "rill.runtime.v1.PullTrigger"
	ResourceKindRefreshTrigger string = "rill.runtime.v1.RefreshTrigger"
	ResourceKindBucketPlanner  string = "rill.runtime.v1.BucketPlanner"
)

// errCyclicDependency is set as the error on resources that can't be reconciled due to a cyclic dependency
var errCyclicDependency = errors.New("cannot be reconciled due to cyclic dependency")

// Reconciler implements reconciliation logic for all resources of a specific kind.
// Reconcilers are managed and invoked by a Controller.
type Reconciler interface {
	Close(ctx context.Context) error
	AssignSpec(from, to *runtimev1.Resource) error
	AssignState(from, to *runtimev1.Resource) error
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
	mu          sync.RWMutex
	running     atomic.Bool
	reconcilers map[string]Reconciler
	catalog     *catalogCache
	// subscribers tracks subscribers to catalog events.
	subscribers map[string]chan map[string]catalogEvent
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
func NewController(ctx context.Context, rt *Runtime, instanceID string, logger *zap.Logger) (*Controller, error) {
	c := &Controller{
		Runtime:        rt,
		InstanceID:     instanceID,
		reconcilers:    make(map[string]Reconciler),
		subscribers:    make(map[string]chan map[string]catalogEvent),
		queue:          make(map[string]*runtimev1.ResourceName),
		queueUpdatedCh: make(chan struct{}),
		timeline:       schedule.New[string, *runtimev1.ResourceName](nameStr),
		invocations:    make(map[string]*invocation),
		completed:      make(chan *invocation),
	}

	// TODO: Setup the logger to duplicate logs to a) the Zap logger, b) an in-memory buffer that exposes the logs over the API
	logger = logger.With(zap.String("instance_id", instanceID))
	c.Logger = slog.New(zapslog.NewHandler(logger.Core()))

	cc, err := newCatalogCache(ctx, c, instanceID)
	if err != nil {
		return nil, err
	}
	c.catalog = cc

	return c, nil
}

// Run starts and runs the controller's event loop. It returns when ctx is cancelled or an unrecoverable error occurs.
// The event loop schedules/invokes resource reconciliation and periodically flushes catalog changes to persistent storage.
// The implementation centers around these internal functions: enqueue, processQueue, schedule, and processCompletedInvocation.
// See their docstrings for further details.
func (c *Controller) Run(ctx context.Context) error {
	if c.running.Swap(true) {
		panic("controller is already running")
	}

	// Check we are still the leader
	err := c.catalog.checkLeader(ctx)
	if err != nil {
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

		timelineTimer.Reset(d)
	}

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
			c.mu.Unlock()
		case inv := <-c.completed: // A reconciler invocation has completed
			c.mu.Lock()
			err = c.processCompletedInvocation(inv)
			if err != nil {
				loopErr = err
				stop = true
			}
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
			c.mu.Lock()
			for _, ch := range c.subscribers {
				ch <- c.catalog.events
			}
			c.catalog.resetEvents()
			c.mu.Unlock()
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
	c.mu.Lock()
	for _, inv := range c.invocations {
		inv.cancel(false)
	}
	c.mu.Unlock()

	// Allow 10 seconds for closing invocations and reconcilers
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wait for all invocations to finish
	for _, inv := range c.invocations {
		select {
		case <-inv.done:
			continue
		case <-ctx.Done():
			closeErr = fmt.Errorf("timed out waiting for reconcile to finish for resource %q", nameStr(inv.name))
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

	c.running.Store(false)
	return closeErr
}

// Get returns a resource by name.
// Soft-deleted resources (i.e. resources where DeletedOn != nil) are not returned.
func (c *Controller) Get(ctx context.Context, name *runtimev1.ResourceName) (*runtimev1.Resource, error) {
	c.checkRunning()
	c.lock(ctx, true)
	defer c.unlock(ctx, true)
	return c.catalog.get(name, false)
}

// List returns a list of resources of the specified kind.
// If kind is empty, all resources are returned.
// Soft-deleted resources (i.e. resources where DeletedOn != nil) are not returned.
func (c *Controller) List(ctx context.Context, kind string) ([]*runtimev1.Resource, error) {
	c.checkRunning()
	c.lock(ctx, true)
	defer c.unlock(ctx, true)
	return c.catalog.list(kind, false)
}

// SubscribeCallback is the callback type passed to Subscribe.
type SubscribeCallback func(e runtimev1.ResourceEvent, n *runtimev1.ResourceName, r *runtimev1.Resource)

// Subscribe registers a callback that will receive resource update events.
func (c *Controller) Subscribe(ctx context.Context, fn SubscribeCallback) error {
	c.checkRunning()

	id := fmt.Sprintf("%v", fn)
	ch := make(chan map[string]catalogEvent)

	c.mu.Lock()
	c.subscribers[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.subscribers, id)
		c.mu.Unlock()
	}()

	for {
		select {
		case events := <-ch:
			for _, e := range events {
				fn(e.event, e.name, e.resource)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Create creates a resource and enqueues it for reconciliation.
// If a resource with the same name is currently being deleted, the deletion will be cancelled.
func (c *Controller) Create(ctx context.Context, name *runtimev1.ResourceName, refs []*runtimev1.ResourceName, owner *runtimev1.ResourceName, paths []string, r *runtimev1.Resource) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	// A deleted resource with the same name may exist and be running. If so, we first cancel it.
	requeued := false
	if inv, ok := c.invocations[nameStr(name)]; ok {
		r, err := c.catalog.get(name, true)
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

	err := c.catalog.create(name, refs, owner, paths, r)
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
	c.checkRunning()
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
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	if !c.isReconcilerForResource(ctx, name) {
		c.cancelIfRunning(name, false)
		c.enqueue(name)
	}

	r, err := c.catalog.get(name, true)
	if err != nil {
		return err
	}

	err = c.safeRename(name, newName)
	if err != nil {
		return err
	}

	err = c.catalog.updateMeta(newName, r.Meta.Refs, owner, paths)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSpec updates a resource's spec and enqueues it for reconciliation.
// If called from outside the resource's reconciler and the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) UpdateSpec(ctx context.Context, name *runtimev1.ResourceName, r *runtimev1.Resource) error {
	c.checkRunning()
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
	c.checkRunning()
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
	c.checkRunning()
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
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	c.cancelIfRunning(name, false)

	if c.isReconcilerForResource(ctx, name) {
		return c.catalog.delete(name)
	}

	err := c.catalog.clearRenamedFrom(name) // Avoid resource being marked both deleted and renamed
	if err != nil {
		return err
	}

	err = c.catalog.updateDeleted(name)
	if err != nil {
		return err
	}

	c.enqueue(name)

	return nil
}

// Flush forces a flush of the controller's catalog changes to persistent storage.
func (c *Controller) Flush(ctx context.Context) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)

	return c.catalog.flush(ctx)
}

// Reconcile enqueues a resource for reconciliation.
// If the resource is currently reconciling, the current reconciler will be cancelled first.
func (c *Controller) Reconcile(ctx context.Context, name *runtimev1.ResourceName) error {
	c.checkRunning()
	c.lock(ctx, false)
	defer c.unlock(ctx, false)
	c.cancelIfRunning(name, false)
	c.enqueue(name)
	return nil
}

// Cancel cancels the current invocation of a resource's reconciler (if it's running).
// It does not re-enqueue the resource for reconciliation.
func (c *Controller) Cancel(ctx context.Context, name *runtimev1.ResourceName) error {
	c.checkRunning()
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
func (c *Controller) checkRunning() {
	if !c.running.Load() {
		panic("controller is not running")
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
	return proto.Equal(inv.name, n)
}

// safeMutateRenamed makes it safe to mutate a resource that's currently being renamed by changing the rename to a delete+create.
// It does nothing if the resource is not currently being renamed (RenamedFrom == nil).
// It must be called while c.mu is held.
func (c *Controller) safeMutateRenamed(n *runtimev1.ResourceName) error {
	r, err := c.catalog.get(n, true)
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

	_, err = c.catalog.get(renamedFrom, true)
	if err == nil {
		// A new resource with the name of the old one has been created in the mean time, so no delete is necessary (reconciler will bring to desired state).
		return nil
	}

	// Create a new resource with the old name, so we can delete it separately.
	err = c.catalog.create(renamedFrom, r.Meta.Refs, r.Meta.Owner, r.Meta.FilePaths, r)
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
	// Just to be safe
	if proto.Equal(from, to) {
		return nil
	}

	// There's a collision if to matches RenamedFrom of another resource.
	collision := false
	for _, n := range c.catalog.renamed {
		r, err := c.catalog.get(n, true)
		if err != nil {
			return fmt.Errorf("internal: failed to get renamed resource %v: %w", n, err)
		}
		if proto.Equal(to, r.Meta.RenamedFrom) {
			collision = true
			break
		}
	}

	// No collision, do a normal rename
	if !collision {
		return c.catalog.rename(from, to)
	}

	// Collision, do a create+delete instead of a rename
	// (since creation might fail if the name is taken, whereas the delete is almost certain to succeed)
	r, err := c.catalog.get(from, true)
	if err != nil {
		return err
	}

	err = c.catalog.create(to, r.Meta.Refs, r.Meta.Owner, r.Meta.FilePaths, r)
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

// processQueue calls schedule() for each resource in c.queue. It is invoked in each iteration of the event loop.
// The reason we have the queue and process it from the event loop (instead of calling schedule() directly from enqueue()) is to enable batching of schedule calls during initialization and when Lock/Unlock is used.
// Batching makes it easier to ensure that parents are scheduled before children when both need to be scheduled.
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
// It returns true if it already now knows that the resource can't be scheduled and will be re-triggered later (e.g. by being added to a waitlist).
// It must be called while c.mu is held.
func (c *Controller) markPending(n *runtimev1.ResourceName) (bool, error) {
	// Remove from timeline (if present)
	c.timeline.Remove(n)

	// Get resource
	r, err := c.catalog.get(n, true)
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

	// Not running - mark pending
	err = c.catalog.updateStatus(n, runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING, time.Time{})
	if err != nil {
		return false, err
	}

	// If resource is cyclic, set error and skip it
	_, cyclic := c.catalog.cyclic[nameStr(n)]
	if cyclic {
		err = c.catalog.updateError(n, errCyclicDependency)
		if err != nil {
			return false, err
		}
		err = c.catalog.updateStatus(n, runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE, time.Time{})
		if err != nil {
			return false, err
		}
		return true, nil
	}

	// Ensure all descendents get marked pending and cancel any running descendents.
	descendentRunning := false
	err = c.catalog.dag.Visit(n, func(ds string, dn *runtimev1.ResourceName) error {
		dr, err := c.catalog.get(dn, true)
		if err != nil {
			return fmt.Errorf("error getting dag node %q: %w", ds, err)
		}
		switch dr.Meta.ReconcileStatus {
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE:
			// Mark it pending
			err = c.catalog.updateStatus(dn, runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING, time.Time{})
			if err != nil {
				return fmt.Errorf("error updating dag node %q: %w", ds, err)
			}
			return nil
		case runtimev1.ReconcileStatus_RECONCILE_STATUS_PENDING:
			// If it's pending, we know all its descendents are also pending.
			// We still need to traverse it to know if any of its descendents are running, but can skip that if we already know a descendent is running (minor optimization).
			if descendentRunning {
				return dag2.ErrSkip
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
			return dag2.ErrSkip // No need to traverse its children - we know they're all pending.
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

	// Proceed to maybeSchedule
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
func (c *Controller) trySchedule(n *runtimev1.ResourceName) (bool, error) {
	r, err := c.catalog.get(n, true)
	if err != nil {
		if errors.Is(err, drivers.ErrResourceNotFound) {
			return true, nil
		}
		return false, err
	}

	// Return true if any parents are pending or running
	parents := c.catalog.dag.Parents(n, true)
	for _, pn := range parents {
		p, err := c.catalog.get(pn, true)
		if err != nil {
			return false, fmt.Errorf("internal: error getting present parent %q: %w", nameStr(pn), err)
		}
		if p.Meta.ReconcileStatus != runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE {
			// When the parent has completed running, processCimpletedInvocation will enqueue its children and we'll run again.
			return true, nil
		}
	}

	// We want deletes to run before renames or regular reconciles.
	// Return false if there are pending deletes and this isn't one of them.
	if len(c.catalog.deleted) != 0 && r.Meta.DeletedOn == nil {
		return false, nil
	}

	// We want renames to run before regular reconciles.
	// Return false if there are pending renames and this isn't one of them.
	if len(c.catalog.renamed) != 0 && r.Meta.RenamedFrom == nil {
		return false, nil
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
		name:     n,
		isDelete: r.Meta.DeletedOn != nil,
		isRename: r.Meta.RenamedFrom != nil,
		done:     make(chan struct{}),
		cancelFn: cancel,
	}
	c.invocations[nameStr(n)] = inv

	// Start reconcile in background
	ctx = contextWithInvocation(ctx, inv)
	reconciler := c.reconciler(n.Kind) // fetched outside of goroutine to keep access under mutex
	go func() {
		defer func() {
			// Catch panics and set as error
			if r := recover(); r != nil {
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
	r, err := c.catalog.get(inv.name, true)
	if err != nil {
		return err
	}

	if inv.isDelete {
		// Trigger processQueue - there may be items in the queue waiting for all deletes to finish
		if len(c.catalog.deleted) == 0 {
			c.setQueueUpdated()
		}

		// Extra checks in case item was re-created during deletion or deleted during a normal reconciling
		if r.Meta.DeletedOn != nil && !inv.cancelled {
			err = c.catalog.delete(r.Meta.Name)
			if err != nil {
				return err
			}
			if inv.result.Err != nil {
				c.Logger.Error("got error while deleting resource", slog.String("name", nameStr(r.Meta.Name)), slog.Any("error", inv.result.Err))
			}
			return nil
		}
	}

	if inv.isRename {
		// Trigger processQueue - there may be items in the queue waiting for all renames to finish
		if len(c.catalog.renamed) == 0 {
			c.setQueueUpdated()
		}

		// Extra checks in case item was cancelled during renaming
		if r.Meta.RenamedFrom != nil && !inv.cancelled {
			err = c.catalog.clearRenamedFrom(r.Meta.Name)
			if err != nil {
				return err
			}
		}
	}

	// Track retrigger time
	if !inv.result.Retrigger.IsZero() {
		if inv.result.Retrigger.After(time.Now()) {
			c.timeline.Set(inv.name, inv.result.Retrigger)
		} else {
			// If retrigger requested before now, we'll just reschedule directly
			inv.reschedule = true
			inv.result.Retrigger = time.Time{}
		}
	}

	// Update status and error
	err = c.catalog.updateStatus(inv.name, runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE, inv.result.Retrigger)
	if err != nil {
		return err
	}
	err = c.catalog.updateError(inv.name, inv.result.Err)
	if err != nil {
		return err
	}

	// Cleanup
	close(inv.done)
	delete(c.invocations, nameStr(inv.name))

	// Let the dominos fall
	if inv.reschedule {
		c.enqueue(inv.name)
	}

	// Enqueue items from waitlist that haven't been updated (and hence re-triggered in the meantime).
	for _, e := range inv.waitlist {
		r, err := c.catalog.get(e.name, true)
		if err != nil {
			if errors.Is(err, drivers.ErrResourceNotFound) {
				continue
			}
			return err
		}
		if r.Meta.SpecVersion == e.specVersion {
			c.enqueue(e.name)
		}
	}

	// When a parent is done, we enqueue its children (unless rescheduling since the children would be blocked anyway)
	if !inv.reschedule {
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
	name       *runtimev1.ResourceName
	isDelete   bool
	isRename   bool
	done       chan struct{}
	cancelFn   context.CancelFunc
	cancelled  bool
	reschedule bool
	holdsLock  bool
	waitlist   map[string]waitlistEntry
	result     ReconcileResult
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
