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
		err := repo.Delete(ctx, path, false)
		require.NoError(t, err)
	}
}

func ReconcileParserAndWait(t testing.TB, rt *runtime.Runtime, id string) {
	ReconcileAndWait(t, rt, id, runtime.GlobalProjectParserName)
}

func ReconcileAndWait(t testing.TB, rt *runtime.Runtime, id string, n *runtimev1.ResourceName) {
	ctrl, err := rt.Controller(context.Background(), id)
	require.NoError(t, err)

	err = ctrl.Reconcile(context.Background(), n)
	require.NoError(t, err)

	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)
}

func RefreshAndWait(t testing.TB, rt *runtime.Runtime, id string, n *runtimev1.ResourceName) {
	ctrl, err := rt.Controller(context.Background(), id)
	require.NoError(t, err)

	// Get resource before refresh
	rPrev, err := ctrl.Get(context.Background(), n, false)
	require.NoError(t, err)

	// Create refresh trigger
	trgName := &runtimev1.ResourceName{Kind: runtime.ResourceKindRefreshTrigger, Name: time.Now().String()}
	err = ctrl.Create(context.Background(), trgName, nil, nil, nil, false, &runtimev1.Resource{
		Resource: &runtimev1.Resource_RefreshTrigger{
			RefreshTrigger: &runtimev1.RefreshTrigger{
				Spec: &runtimev1.RefreshTriggerSpec{
					Resources: []*runtimev1.ResourceName{n},
				},
			},
		},
	})
	require.NoError(t, err)

	// Wait for refresh to complete
	err = ctrl.WaitUntilIdle(context.Background(), false)
	require.NoError(t, err)

	// Get resource after refresh
	rNew, err := ctrl.Get(context.Background(), n, false)
	require.NoError(t, err)

	// Check the resource's spec version has increased
	require.Greater(t, rNew.Meta.SpecVersion, rPrev.Meta.SpecVersion)
}

func RequireReconcileState(t testing.TB, rt *runtime.Runtime, id string, lenResources, lenReconcileErrs, lenParseErrs int) {
	ctrl, err := rt.Controller(context.Background(), id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", "", false)
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

	var names []string
	for _, r := range rs {
		names = append(names, fmt.Sprintf("%s/%s", r.Meta.Name.Kind, r.Meta.Name.Name))
	}

	if lenParseErrs >= 0 {
		require.Equal(t, lenParseErrs, len(parseErrs), "parse errors: %s", strings.Join(parseErrs, "\n"))
	}
	if lenReconcileErrs >= 0 {
		require.Equal(t, lenReconcileErrs, len(reconcileErrs), "reconcile errors: %s", strings.Join(reconcileErrs, "\n"))
	}
	if lenResources >= 0 {
		require.Equal(t, lenResources, len(rs), "resources: %s", strings.Join(names, "\n"))
	}
}

func GetResource(t testing.TB, rt *runtime.Runtime, id, kind, name string) *runtimev1.Resource {
	ctrl, err := rt.Controller(context.Background(), id)
	require.NoError(t, err)

	r, err := ctrl.Get(context.Background(), &runtimev1.ResourceName{Kind: kind, Name: name}, true)
	require.NoError(t, err)

	return r
}

func RequireResource(t testing.TB, rt *runtime.Runtime, id string, a *runtimev1.Resource) {
	ctrl, err := rt.Controller(context.Background(), id)
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
		state.RefsHash = ""
	case runtime.ResourceKindAlert:
		state := b.GetAlert().State
		state.SpecHash = ""
		state.RefsHash = ""
		for i, e := range state.ExecutionHistory {
			e.StartedOn = nil
			e.FinishedOn = nil
			if a.GetAlert().State.ExecutionHistory[i].ExecutionTime == nil {
				e.ExecutionTime = nil
			}
		}
	case runtime.ResourceKindConnector:
		state := b.GetConnector().State
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
	ctrl, err := rt.Controller(context.Background(), id)
	require.NoError(t, err)

	rs, err := ctrl.List(context.Background(), "", "", false)
	require.NoError(t, err)

	for _, r := range rs {
		t.Logf("%s/%s: status=%d, stateversion=%d, error=%q", r.Meta.Name.Kind, r.Meta.Name.Name, r.Meta.ReconcileStatus, r.Meta.StateVersion, r.Meta.ReconcileError)
	}
}

func RequireParseErrors(t testing.TB, rt *runtime.Runtime, id string, expectedParseErrors map[string]string) {
	ctrl, err := rt.Controller(context.Background(), id)
	require.NoError(t, err)

	pp, err := ctrl.Get(context.Background(), runtime.GlobalProjectParserName, true)
	require.NoError(t, err)

	parseErrs := map[string]string{}
	for _, pe := range pp.GetProjectParser().State.ParseErrors {
		parseErrs[pe.FilePath] = pe.Message
	}
	require.Len(t, parseErrs, len(expectedParseErrors), "Should have %d parse errors", len(expectedParseErrors))

	for f, pe := range parseErrs {
		// Checking parseError using Contains instead of Equal
		require.Contains(t, pe, expectedParseErrors[f])
	}
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
