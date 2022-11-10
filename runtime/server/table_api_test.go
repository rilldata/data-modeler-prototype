package server

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/rilldata/rill/runtime/api"
	"github.com/rilldata/rill/runtime/drivers"
)

func getSingleValue(t *testing.T, rows *sqlx.Rows) int {
	var val int
	for rows.Next() {
		err := rows.Scan(&val)
		require.NoError(t, err)
	}
	rows.Close()
	return val
}

func TestServer_Database(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	result := createTestTable(server, instanceId, t)
	require.Equal(t, 1, getSingleValue(t, result.Rows))
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	require.NoError(t, err)
	require.Equal(t, 1, getSingleValue(t, result.Rows))
}

func createTestTable(server *Server, instanceId string, t *testing.T) *drivers.Result {
	result, err := server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "create table test (a int)",
	})
	require.NoError(t, err)
	result.Close()
	result, _ = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "insert into test values (1)",
	})
	result.Close()
	result, err = server.query(context.Background(), instanceId, &drivers.Statement{
		Query: "select count(*) from test",
	})
	require.NoError(t, err)
	return result
}

func TestServer_TableCardinality(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.TableCardinality(context.Background(), &api.CardinalityRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), cr.Cardinality)
}

func TestServer_ProfileColumns(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.ProfileColumns(context.Background(), &api.ProfileColumnsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 0, len(cr.ProfileColumn))
}

func TestServer_TableRows(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	cr, err := server.TableRows(context.Background(), &api.RowsRequest{
		InstanceId: instanceId,
		TableName:  "test",
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(cr.Data))
}

func TestServer_RenameObject(t *testing.T) {
	server, instanceId, err := getTestServer(t)
	require.NoError(t, err)
	rows := createTestTable(server, instanceId, t)
	rows.Close()
	_, err = server.RenameDatabaseObject(context.Background(), &api.RenameDatabaseObjectRequest{
		InstanceId: instanceId,
		Name:       "test",
		Newname:    "test2",
		Type:       api.DatabaseObjectType_TABLE.Enum(),
	})
	require.NoError(t, err)
}
