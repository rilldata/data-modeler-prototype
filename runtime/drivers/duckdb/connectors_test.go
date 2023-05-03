package duckdb

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	_ "github.com/rilldata/rill/runtime/connectors/gcs"
	_ "github.com/rilldata/rill/runtime/connectors/s3"
)

func TestConnectorWithSourceVariations(t *testing.T) {
	testdataPathRel := "../../../web-local/test/data"
	testdataPathAbs, err := filepath.Abs(testdataPathRel)
	require.NoError(t, err)

	sources := []struct {
		Connector       string
		Path            string
		AdditionalProps map[string]any
	}{
		{"local_file", filepath.Join(testdataPathRel, "AdBids.csv"), nil},
		{"local_file", filepath.Join(testdataPathRel, "AdBids.csv"), map[string]any{"csv.delimiter": ","}},
		{"local_file", filepath.Join(testdataPathRel, "AdBids.csv.gz"), nil},
		{"local_file", filepath.Join(testdataPathRel, "AdBids.parquet"), map[string]any{"hive_partitioning": true}},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.parquet"), nil},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.txt"), nil},
		{"local_file", "../../../runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz", nil},
		// something wrong with this particular file. duckdb fails to extract
		// TODO: move the generator to go and fix the parquet file
		//{"local_file", testdataPath + "AdBids.parquet.gz", nil},
		// only enable to do adhoc tests. needs credentials to work
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.csv.gz", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet", nil},
		//{"s3", "s3://rill-developer.rilldata.io/AdBids.parquet.gz", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.csv.gz", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet", nil},
		//{"gcs", "gs://scratch.rilldata.com/rill-developer/AdBids.parquet.gz", nil},
	}

	ctx := context.Background()
	conn, err := Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			var props map[string]any
			if tt.AdditionalProps != nil {
				props = tt.AdditionalProps
			} else {
				props = make(map[string]any)
			}
			props["path"] = tt.Path

			e := &connectors.Env{
				RepoDriver:      "file",
				RepoRoot:        ".",
				AllowHostAccess: true,
			}
			s := &connectors.Source{
				Name:       "foo",
				Connector:  tt.Connector,
				Properties: props,
			}
			_, err = olap.Ingest(ctx, e, s)
			require.NoError(t, err)

			var count int
			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.GreaterOrEqual(t, count, 100)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestConnectorWithGithubRepoDriver(t *testing.T) {
	testdataPathRel := "../../../web-local/test/data"
	testdataPathAbs, err := filepath.Abs(testdataPathRel)
	require.NoError(t, err)

	sources := []struct {
		Connector string
		Path      string
		repoRoot  string
		isError   bool
	}{
		{"local_file", "AdBids.csv", testdataPathAbs, false},
		{"local_file", filepath.Join(testdataPathAbs, "AdBids.csv"), testdataPathAbs, false},
		{"local_file", "../../../runtime/testruntime/testdata/ad_bids/data/AdBids.csv.gz", testdataPathAbs, true},
	}

	ctx := context.Background()
	conn, err := Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	for _, tt := range sources {
		t.Run(fmt.Sprintf("%s - %s", tt.Connector, tt.Path), func(t *testing.T) {
			props := make(map[string]any)
			props["path"] = tt.Path

			e := &connectors.Env{
				RepoDriver:      "github",
				RepoRoot:        tt.repoRoot,
				AllowHostAccess: false,
			}
			s := &connectors.Source{
				Name:       "foo",
				Connector:  tt.Connector,
				Properties: props,
			}
			_, err = olap.Ingest(ctx, e, s)
			if tt.isError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var count int
			rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
			require.NoError(t, err)
			require.True(t, rows.Next())
			require.NoError(t, rows.Scan(&count))
			require.GreaterOrEqual(t, count, 100)
			require.False(t, rows.Next())
			require.NoError(t, rows.Close())
		})
	}
}

func TestCSVDelimiter(t *testing.T) {
	ctx := context.Background()
	conn, err := Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	testdataPathAbs, err := filepath.Abs("../../../web-local/test/data")
	require.NoError(t, err)
	testDelimiterCsvPath := filepath.Join(testdataPathAbs, "test-delimiter.csv")

	_, err = olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "foo",
		Connector: "local_file",
		Properties: map[string]any{
			"path": testDelimiterCsvPath,
		},
	})
	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	// 3 columns because no delimiter is passed
	require.Len(t, cols, 3)
	require.NoError(t, rows.Close())

	_, err = olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "foo",
		Connector: "local_file",
		Properties: map[string]any{
			"path":   testDelimiterCsvPath,
			"duckdb": map[string]any{"delim": "'+'"},
		},
	})
	require.NoError(t, err)
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
	require.NoError(t, err)
	cols, err = rows.Columns()
	require.NoError(t, err)
	// 3 columns because no delimiter is passed
	require.Len(t, cols, 2)
	require.NoError(t, rows.Close())
}

func TestFileFormatAndDelimiter(t *testing.T) {
	ctx := context.Background()
	conn, err := Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, _ := conn.OLAPStore()

	testdataPathAbs, err := filepath.Abs("../../../web-local/test/data")
	require.NoError(t, err)
	testDelimiterCsvPath := filepath.Join(testdataPathAbs, "test-format.log")

	_, err = olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "foo",
		Connector: "local_file",
		Properties: map[string]any{
			"path":   testDelimiterCsvPath,
			"format": "csv",
			"duckdb": map[string]any{"delim": "' '"},
		},
	})
	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM foo"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	// 5 columns in file
	require.Len(t, cols, 5)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(timestamp) FROM foo"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 8)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestCSVIngestionWithColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.csv")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "csv_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"auto_detect":   false,
				"header":        true,
				"ignore_errors": true,
				"columns":       "{id:'INTEGER',name:'VARCHAR',country:'VARCHAR',city:'VARCHAR'}",
			},
		},
	})
	require.NoError(t, err)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM csv_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 4)
	require.ElementsMatch(t, cols, [4]string{"id", "name", "country", "city"})
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM csv_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 100)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionDefault(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
		},
	})
	require.NoError(t, err)

	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 9)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"columns": "{id:'INTEGER', name:'VARCHAR', isActive:'BOOLEAN', createdDate:'VARCHAR', address:'STRUCT(street VARCHAR, city VARCHAR, postalCode VARCHAR)', tags:'VARCHAR[]', projects:'STRUCT(projectId INTEGER, projectName VARCHAR, startDate VARCHAR, endDate VARCHAR)[]', scores:'INTEGER[]'}",
			},
		},
	})
	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 8)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithLessColumns(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"columns": "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR',}",
			},
		},
	})
	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 4)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithVariousParams(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"maximum_object_size": "9999999",
				"lines":               true,
				"ignore_errors":       true,
				"compression":         "auto",
				"columns":             "{id:'INTEGER',name:'VARCHAR',isActive:'BOOLEAN',createdDate:'VARCHAR',}",
				"json_format":         "records",
				"auto_detect":         false,
				"sample_size":         -1,
				"dateformat":          "iso",
				"timestampformat":     "iso",
			},
		},
	})
	require.NoError(t, err)
	rows, err := olap.Execute(ctx, &drivers.Statement{Query: "SELECT * FROM json_source"})
	require.NoError(t, err)
	cols, err := rows.Columns()
	require.NoError(t, err)
	require.Len(t, cols, 4)
	require.NoError(t, rows.Close())

	var count int
	rows, err = olap.Execute(ctx, &drivers.Statement{Query: "SELECT count(*) FROM json_source"})
	require.NoError(t, err)
	require.True(t, rows.Next())
	require.NoError(t, rows.Scan(&count))
	require.Equal(t, count, 10)
	require.False(t, rows.Next())
	require.NoError(t, rows.Close())
}

func TestJsonIngestionWithInvalidParam(t *testing.T) {
	olap := runOLAPStore(t)
	ctx := context.Background()
	filePath := createFilePath(t, "../../../web-local/test/data", "Users.json")

	_, err := olap.Ingest(ctx, &connectors.Env{
		RepoDriver:      "file",
		RepoRoot:        ".",
		AllowHostAccess: true,
	}, &connectors.Source{
		Name:      "json_source",
		Connector: "local_file",
		Properties: map[string]any{
			"path": filePath,
			"duckdb": map[string]any{
				"json": map[string]any{
					"invalid_param": "auto",
				},
			},
		},
	})
	require.Error(t, err, "Invalid named parameter \"invalid_param\" for function read_json")
}

func createFilePath(t *testing.T, dirPath string, fileName string) string {
	testdataPathAbs, err := filepath.Abs(dirPath)
	require.NoError(t, err)
	filePath := filepath.Join(testdataPathAbs, fileName)
	return filePath
}

func runOLAPStore(t *testing.T) drivers.OLAPStore {
	conn, err := Driver{}.Open("?access_mode=read_write", zap.NewNop())
	require.NoError(t, err)
	olap, canServe := conn.OLAPStore()
	require.True(t, canServe)
	return olap
}
