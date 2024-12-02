package duckdb

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	activity "github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestConfig(t *testing.T) {
	cfg, err := newConfig(map[string]any{}, "")
	require.NoError(t, err)
	require.Equal(t, "?custom_user_agent=rill", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": ":memory:?memory_limit=2GB"}, "")
	require.NoError(t, err)
	require.Equal(t, "?custom_user_agent=rill&memory_limit=2GB", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "", "memory_limit_gb": "1", "cpu": 2}, "")
	require.NoError(t, err)
	require.Equal(t, "?custom_user_agent=rill&max_memory=1GB&threads=2", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)
	require.Equal(t, true, cfg.ExtTableStorage)

	cfg, err = newConfig(map[string]any{}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "path/to/main.db?custom_user_agent=rill", cfg.DSN)
	require.Equal(t, "path/to/main.db", cfg.DBFilePath)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"pool_size": 10}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "path/to/main.db?custom_user_agent=rill", cfg.DSN)
	require.Equal(t, "path/to/main.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"pool_size": "10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "?rill_pool_size=4", "pool_size": "10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, 4, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?rill_pool_size=10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?custom_user_agent=rill", cfg.DSN)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)
	require.Equal(t, 10, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB&rill_pool_size=10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "path/to/duck.db?custom_user_agent=rill&max_memory=4GB", cfg.DSN)
	require.Equal(t, 10, cfg.PoolSize)
	require.Equal(t, "path/to/duck.db", cfg.DBFilePath)

	_, err = newConfig(map[string]any{"dsn": "path/to/duck.db?max_memory=4GB", "pool_size": "abc"}, "path/to")
	require.Error(t, err)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db?rill_pool_size=10"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db", "memory_limit_gb": "8", "cpu": "2"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)
	require.Equal(t, "duck.db?custom_user_agent=rill&max_memory=8GB&threads=2", cfg.DSN)
	require.Equal(t, 2, cfg.PoolSize)

	cfg, err = newConfig(map[string]any{"dsn": "duck.db?max_memory=2GB&rill_pool_size=4"}, "path/to")
	require.NoError(t, err)
	require.Equal(t, "duck.db", cfg.DBFilePath)
	require.Equal(t, "duck.db?custom_user_agent=rill&max_memory=2GB", cfg.DSN)
	require.Equal(t, 4, cfg.PoolSize)
}

func Test_specialCharInPath(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "let's t@st \"weird\" dirs")
	err := os.Mkdir(path, fs.ModePerm)
	require.NoError(t, err)

	dbFile := filepath.Join(path, "st@g3's.db")
	conn, err := Driver{}.Open("default", map[string]any{"path": dbFile, "memory_limit_gb": "4", "cpu": "1", "external_table_storage": false}, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	config := conn.(*connection).config
	require.Equal(t, filepath.Join(path, "st@g3's.db?custom_user_agent=rill&max_memory=4GB&threads=1"), config.DSN)
	require.Equal(t, 2, config.PoolSize)

	olap, ok := conn.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT 1"})
	require.NoError(t, err)
	require.NoError(t, res.Close())
	require.NoError(t, conn.Close())
}

func TestOverrides(t *testing.T) {
	cfgMap := map[string]any{"path": "duck.db", "memory_limit_gb": "4", "cpu": "2", "max_memory_gb_override": "2", "threads_override": "10", "external_table_storage": false}
	handle, err := Driver{}.Open("default", cfgMap, storage.MustNew(t.TempDir(), nil), activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)

	olap, ok := handle.AsOLAP("")
	require.True(t, ok)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT value FROM duckdb_settings() WHERE name='max_memory'"})
	require.NoError(t, err)
	require.True(t, res.Next())
	var mem string
	require.NoError(t, res.Scan(&mem))
	require.NoError(t, res.Close())

	require.Equal(t, "1.8 GiB", mem)
}
