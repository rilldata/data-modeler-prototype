package testruntime

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func PutFiles(t testing.TB, rt *runtime.Runtime, id string, files map[string]string) {
	ctx := context.Background()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	for path, data := range files {
		err := repo.Put(ctx, path, strings.NewReader(strings.TrimSpace(data)))
		require.NoError(t, err)
	}
}

func RenameFile(t testing.TB, rt *runtime.Runtime, id, from, to string) {
	ctx := context.Background()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	require.NoError(t, repo.Rename(ctx, from, to))
}

func DeleteFiles(t testing.TB, rt *runtime.Runtime, id string, files ...string) {
	ctx := context.Background()
	repo, release, err := rt.Repo(ctx, id)
	require.NoError(t, err)
	defer release()

	for _, path := range files {
		err := repo.Delete(ctx, path)
		require.NoError(t, err)
	}
}

func ReconcileParserAndWait(t testing.TB, rt *runtime.Runtime, id string) {
	ReconcileAndWait(t, rt, id, runtime.GlobalProjectParserName)
}

func ReconcileAndWait(t testing.TB, rt *runtime.Runtime, id string, n *runtimev1.ResourceName) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	err = ctrl.Reconcile(context.Background(), n)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)
}

func RefreshAndWait(t testing.TB, rt *runtime.Runtime, id string, n *runtimev1.ResourceName) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	// Get spec version before refresh
	r, err := ctrl.Get(context.Background(), n, false)
	require.NoError(t, err)
	prevSpecVersion := r.Meta.SpecVersion

	// Create refresh trigger
	trgName := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: time.Now().String()}
	err = ctrl.Create(context.Background(), trgName, nil, nil, nil, true, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					OnlyNames: []*runtimev1.ResourceName{n},
				},
			},
		},
	})
	require.NoError(t, err)

	// Wait for refresh to complete
	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)

	// Check the resource's spec version has increased
	require.Greater(t, r.Meta.SpecVersion, prevSpecVersion)
}

func WaitUntilIdle(t testing.TB, rt *runtime.Runtime, id string) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	// Smaller times was not stable on an M1 mac.
	// TODO: Refactor to wait for the controller to actually be triggered if we ever have instability
	time.Sleep(time.Second)

	// For now this is only used for continuous watcher. add ignore hidden to param if otherwise
	err = ctrl.WaitUntilIdle(context.Background(), true)
	require.NoError(t, err)
}

func RequireReconcileState(t testing.TB, rt *runtime.Runtime, id string, lenResources, lenReconcileErrs, lenParseErrs int) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", false)
	require.NoError(t, err)

	var reconcileErrs, parseErrs []string
	for _, r := range rs {
		if r.Meta.ReconcileError != "" {
			reconcileErrs = append(reconcileErrs, fmt.Sprintf("%s/%s: %s", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileError))
		}

		if r.Meta.Name.Kind == runtime.ResourceKindProjectParser {
			for _, pe := range r.GetProjectParser().State.ParseErrors {
				parseErrs = append(parseErrs, fmt.Sprintf("%s: %s", pe.FilePath, pe.Message))
			}
		}
	}

	require.Equal(t, lenParseErrs, len(parseErrs), "parse errors: %s", strings.Join(parseErrs, "\n"))
	require.Equal(t, lenReconcileErrs, len(reconcileErrs), "reconcile errors: %s", strings.Join(reconcileErrs, "\n"))
	require.Equal(t, lenResources, len(rs), "resources")
}

func RequireResource(t testing.TB, rt *runtime.Runtime, id string, a *runtimev1.Resource) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	b, err := ctrl.Get(context.Background(), a.Meta.Name, true) // Set clone=true because we may manipulate it before comparing
	require.NoError(t, err)

	require.True(t, proto.Equal(a.Meta.Name, b.Meta.Name), "expected: %v\nactual: %v", a.Meta.Name, b.Meta.Name)
	require.ElementsMatch(t, a.Meta.Refs, b.Meta.Refs)
	require.True(t, proto.Equal(a.Meta.Owner, b.Meta.Owner), "expected: %v\nactual: %v", a.Meta.Owner, b.Meta.Owner)
	require.ElementsMatch(t, a.Meta.FilePaths, b.Meta.FilePaths)
	require.Greater(t, b.Meta.Version, int64(0))
	require.Greater(t, b.Meta.SpecVersion, int64(0))
	require.Greater(t, b.Meta.StateVersion, int64(0))
	require.NotEmpty(t, b.Meta.CreatedOn.AsTime())
	require.NotEmpty(t, b.Meta.SpecUpdatedOn.AsTime())
	require.NotEmpty(t, b.Meta.StateUpdatedOn.AsTime())
	require.Nil(t, b.Meta.DeletedOn)

	// Checking ReconcileError using Contains instead of Equal
	if a.Meta.ReconcileError == "" {
		require.Empty(t, b.Meta.ReconcileError)
	} else {
		require.Contains(t, b.Meta.ReconcileError, a.Meta.ReconcileError)
	}

	// Not comparing these fields because they are not stable:
	// require.Equal(t, a.Meta.ReconcileStatus, b.Meta.ReconcileStatus)
	// require.Equal(t, a.Meta.ReconcileOn, b.Meta.ReconcileOn)
	// require.Equal(t, a.Meta.RenamedFrom, b.Meta.RenamedFrom)

	// Some kind-specific fields are not stable. We reset those to stable values before comparing.
	switch b.Meta.Name.Kind {
	case runtime.ResourceKindSource:
		state := b.GetSource().State
		state.RefreshedOn = nil
		state.SpecHash = ""
	case runtime.ResourceKindModel:
		state := b.GetModel().State
		state.RefreshedOn = nil
		state.SpecHash = ""
	}

	// Hack to only compare the Resource field (not Meta)
	name := b.Meta.Name
	a = &runtimev1.Resource{Resource: a.Resource}
	b = &runtimev1.Resource{Resource: b.Resource}

	// Compare!
	require.True(t, proto.Equal(a, b), "for resource %q\nexpected: %v\nactual: %v", name.Name, a.Resource, b.Resource)
}

func DumpResources(t testing.TB, rt *runtime.Runtime, id string) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", false)
	require.NoError(t, err)

	for _, r := range rs {
		t.Logf("%s/%s: status=%d, stateversion=%d, error=%q", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileStatus, r.Meta.StateVersion, r.Meta.ReconcileError)
	}
}

func RequireParseErrors(t testing.TB, rt *runtime.Runtime, id string, expectedParseErrors map[string]string) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	pp, err := ctrl.Get(context.Background(), runtime.GlobalProjectParserName, true)
	require.NoError(t, err)

	parseErrs := map[string]string{}
	for _, pe := range pp.GetProjectParser().State.ParseErrors {
		parseErrs[pe.FilePath] = pe.Message
	}
	require.Len(t, parseErrs, len(expectedParseErrors))

	for f, pe := range parseErrs {
		// Checking parseError using Contains instead of Equal
		require.Contains(t, pe, expectedParseErrors[f])
	}
}

func WaitForResource(t testing.TB, rt *runtime.Runtime, id, name, path string) (*runtimev1.Resource, string) {
	ctrl, err := rt.Controller(id)
	require.NoError(t, err)

	var res *runtimev1.Resource
	errStr := ""

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	// ignore error since cancelling will cause error as well
	_ = ctrl.Subscribe(ctx, func(_ runtimev1.ResourceEvent, n *runtimev1.ResourceName, r *runtimev1.Resource) {
		if n.Name == name && r != nil && r.Meta.ReconcileStatus == runtimev1.ReconcileStatus_RECONCILE_STATUS_IDLE {
			switch r.Resource.(type) {
			case *runtimev1.Resource_Source:
				if r.GetSource().State.Table == "" {
					return
				}

			case *runtimev1.Resource_Model:
				if r.GetModel().State.Table == "" {
					return
				}

			case *runtimev1.Resource_MetricsView:
				if r.GetMetricsView().State.ValidSpec == nil {
					return
				}
			}
			// if the resource is the one we want return
			res = r
			cancel()
		} else if n.Kind == runtime.ResourceKindProjectParser {
			// else check for errors
			for _, parseError := range r.GetProjectParser().State.ParseErrors {
				if parseError.FilePath == path {
					errStr = parseError.Message
					cancel()
					break
				}
			}
		}
	})

	// Wait for the resource to be saved to the db
	// TODO: is there a better way?
	time.Sleep(250 * time.Millisecond)

	return res, errStr
}

func WaitRequireResource(t testing.TB, rt *runtime.Runtime, id string, a *runtimev1.Resource) {
	_, recErr := WaitForResource(t, rt, id, a.Meta.Name.Name, a.Meta.FilePaths[0])
	require.Equal(t, recErr, "", "unexpected parse error")

	RequireResource(t, rt, id, a)
}

func WaitRequireParseError(t testing.TB, rt *runtime.Runtime, id, name, path, err string) {
	_, recErr := WaitForResource(t, rt, id, name, path)
	require.Equal(t, recErr, err)
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
