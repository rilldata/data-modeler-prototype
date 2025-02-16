package reconcilers

import (
	"context"
	"errors"
	"fmt"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"golang.org/x/sync/semaphore"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindSource, newSourceReconciler)
}

type SourceReconciler struct {
	C       *runtime.Controller
	execSem *semaphore.Weighted
}

func newSourceReconciler(ctx context.Context, c *runtime.Controller) (runtime.Reconciler, error) {
	cfg, err := c.Runtime.InstanceConfig(ctx, c.InstanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get model execution concurrency limit: %w", err)
	}
	// Re-using the model limit since we are deprecating sources soon (so everything will be a model).
	if cfg.ModelConcurrentExecutionLimit <= 0 {
		return nil, errors.New("model_concurrent_execution_limit must be greater than zero")
	}
	return &SourceReconciler{
		C:       c,
		execSem: semaphore.NewWeighted(int64(cfg.ModelConcurrentExecutionLimit)),
	}, nil
}

func (r *SourceReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *SourceReconciler) AssignSpec(from, to *runtimev1.Resource) error {
	a := from.GetSource()
	b := to.GetSource()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign spec from %T to %T", from.Resource, to.Resource)
	}
	b.Spec = a.Spec
	return nil
}

func (r *SourceReconciler) AssignState(from, to *runtimev1.Resource) error {
	a := from.GetSource()
	b := to.GetSource()
	if a == nil || b == nil {
		return fmt.Errorf("cannot assign state from %T to %T", from.Resource, to.Resource)
	}
	b.State = a.State
	return nil
}

func (r *SourceReconciler) ResetState(res *runtimev1.Resource) error {
	res.GetSource().State = &runtimev1.SourceState{}
	return nil
}

func (r *SourceReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n, true)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	src := self.GetSource()
	if src == nil {
		return runtime.ReconcileResult{Err: errors.New("not a source")}
	}

	// For existing projects that have sources the parser will delete source and add a model.
	// We want to keep the table so we can keep serving the dashboards hence making the source reconciler a no-op.
	return runtime.ReconcileResult{}
}
