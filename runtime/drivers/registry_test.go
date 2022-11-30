package drivers_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
)

func testRegistry(t *testing.T, reg drivers.RegistryStore) {
	ctx := context.Background()
	inst := &drivers.Instance{
		OLAPDriver:   "duckdb",
		OLAPDSN:      ":memory:",
		RepoDriver:   "file",
		RepoDSN:      ".",
		EmbedCatalog: true,
	}

	err := reg.CreateInstance(ctx, inst)
	require.NoError(t, err)
	_, err = uuid.Parse(inst.ID)
	require.NoError(t, err)
	require.Equal(t, "duckdb", inst.OLAPDriver)
	require.Equal(t, ":memory:", inst.OLAPDSN)
	require.Equal(t, "file", inst.RepoDriver)
	require.Equal(t, ".", inst.RepoDSN)
	require.Equal(t, true, inst.EmbedCatalog)
	require.Greater(t, time.Minute, time.Since(inst.CreatedOn))
	require.Greater(t, time.Minute, time.Since(inst.UpdatedOn))

	res, found, err := reg.FindInstance(ctx, inst.ID)
	require.True(t, found)
	require.Equal(t, inst.OLAPDriver, res.OLAPDriver)
	require.Equal(t, inst.OLAPDSN, res.OLAPDSN)
	require.Equal(t, inst.RepoDriver, res.RepoDriver)
	require.Equal(t, inst.RepoDSN, res.RepoDSN)
	require.Equal(t, inst.EmbedCatalog, res.EmbedCatalog)

	err = reg.CreateInstance(ctx, &drivers.Instance{OLAPDriver: "druid"})
	require.NoError(t, err)

	insts, err := reg.FindInstances(ctx)
	require.Equal(t, 2, len(insts))

	err = reg.DeleteInstance(ctx, inst.ID)
	require.NoError(t, err)

	_, found, err = reg.FindInstance(ctx, inst.ID)
	require.False(t, found)

	insts, err = reg.FindInstances(ctx)
	require.Equal(t, 1, len(insts))
}
