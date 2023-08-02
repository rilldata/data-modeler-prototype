package runtime

import (
	"context"
	"testing"

	"github.com/google/uuid"
	_ "github.com/rilldata/rill/runtime/drivers/sqlite"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConnectionCache(t *testing.T) {
	ctx := context.Background()
	id := uuid.NewString()

	c := newConnectionCache(10, zap.NewNop())
	conn1, release, err := c.get(ctx, id, "sqlite", ":memory:", false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn1)

	conn2, release, err := c.get(ctx, id, "sqlite", ":memory:", false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn2)

	conn3, release, err := c.get(ctx, uuid.NewString(), "sqlite", ":memory:", false)
	require.NoError(t, err)
	release()
	require.NotNil(t, conn3)

	require.True(t, conn1 == conn2)
	require.False(t, conn2 == conn3)
}
