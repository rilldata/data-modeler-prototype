package transporter

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

func Test_sqlextensionToDuckDB_Transfer(t *testing.T) {
	tempDir := t.TempDir()

	dbPath := fmt.Sprintf("%s.db", tempDir)
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	_, err = db.Exec(`
	drop table if exists t;
	create table t(i);
	insert into t values(42), (314);
	`)
	require.NoError(t, err)
	db.Close()

	from, err := drivers.Open("sqlite_ext", map[string]any{"dsn": ""}, false, zap.NewNop())
	require.NoError(t, err)
	to, err := drivers.Open("duckdb", map[string]any{"dsn": ""}, false, zap.NewNop())
	require.NoError(t, err)
	olap, _ := to.AsOLAP("")

	tr := &sqlextensionToDuckDB{
		to:     olap,
		from:   from,
		logger: zap.NewNop(),
	}
	query := fmt.Sprintf("SELECT * FROM sqlite_scan('%s', 't');", dbPath)
	err = tr.Transfer(context.Background(), &drivers.DatabaseSource{SQL: query}, &drivers.DatabaseSink{Table: "test"}, &drivers.TransferOpts{}, drivers.NoOpProgress{})
	require.NoError(t, err)

	res, err := olap.Execute(context.Background(), &drivers.Statement{Query: "SELECT count(*) from test"})
	require.NoError(t, err)
	res.Next()
	var count int
	err = res.Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 2, count)
}
