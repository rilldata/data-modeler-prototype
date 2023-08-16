package reconcilers

import (
	"context"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"google.golang.org/protobuf/proto"
)

func init() {
	runtime.RegisterReconcilerInitializer(runtime.ResourceKindRefreshTrigger, newRefreshTriggerReconciler)
}

// RefreshTriggerReconciler reconciles a RefreshTrigger.
// When a RefreshTrigger is created, the reconciler will refresh source and model by setting Trigger=true in their specs.
// After that, it will delete the RefreshTrigger resource.
type RefreshTriggerReconciler struct {
	C *runtime.Controller
}

func newRefreshTriggerReconciler(c *runtime.Controller) runtime.Reconciler {
	return &RefreshTriggerReconciler{C: c}
}

func (r *RefreshTriggerReconciler) Close(ctx context.Context) error {
	return nil
}

func (r *RefreshTriggerReconciler) Reconcile(ctx context.Context, n *runtimev1.ResourceName) runtime.ReconcileResult {
	self, err := r.C.Get(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}
	trigger := self.GetRefreshTrigger()

	if self.Meta.Deleted {
		return runtime.ReconcileResult{}
	}

	resources, err := r.C.List(ctx)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	for _, res := range resources {
		if len(trigger.Spec.OnlyNames) > 0 {
			found := false
			for _, n := range trigger.Spec.OnlyNames {
				if n.Kind == "" && n.Name == res.Meta.Name.Name || proto.Equal(res.Meta.Name, n) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		updated := true
		switch res.Meta.Name.Kind {
		case runtime.ResourceKindSource:
			source := res.GetSource()
			source.Spec.Trigger = true
		case runtime.ResourceKindModel:
			model := res.GetModel()
			model.Spec.Trigger = true
		default:
			updated = false
		}
		if updated {
			err = r.C.UpdateSpec(ctx, res.Meta.Name, res.Meta.Refs, res.Meta.Owner, res.Meta.FilePaths, res)
			if err != nil {
				return runtime.ReconcileResult{Err: err}
			}
		}
	}

	err = r.C.Delete(ctx, n)
	if err != nil {
		return runtime.ReconcileResult{Err: err}
	}

	return runtime.ReconcileResult{}
}
