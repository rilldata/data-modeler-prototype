package duckdb

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

func (c *connection) Ingest(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	err := source.Validate()
	if err != nil {
		return err
	}

	// Driver-specific overrides
	// switch source.Connector {
	// case "local_file":
	// 	return c.ingestFile(ctx, env, source)
	// }

	if source.Connector == "local_file" {
		return c.ingestLocalFiles(ctx, env, source)
	}

	localPaths, err := connectors.ConsumeAsFiles(ctx, env, source)
	if err != nil {
		return err
	}
	defer fileutil.ForceRemoveFiles(localPaths)
	// multiple parquet files can be loaded in single sql
	// this seems to be performing very fast as compared to appending individual files
	return c.ingestFiles(ctx, source, localPaths)
}

// for files downloaded locally from remote sources
func (c *connection) ingestFiles(ctx context.Context, source *connectors.Source, filenames []string) error {
	config, err := connectors.ParseCommonConfig(source.Properties)
	if err != nil {
		return err
	}

	from, err := getSourceReader(filenames, config)
	if err != nil {
		return err
	}
	query := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s);", source.Name, from)
	return c.Exec(ctx, &drivers.Statement{Query: query, Priority: 1})
}

// local files
func (c *connection) ingestLocalFiles(ctx context.Context, env *connectors.Env, source *connectors.Source) error {
	conf, err := connectors.ParseCommonConfig(source.Properties)
	if err != nil {
		return err
	}

	path := conf.Path
	if !filepath.IsAbs(path) {
		// If the path is relative, it's relative to the repo root
		if env.RepoDriver != "file" || env.RepoDSN == "" {
			return fmt.Errorf("file connector cannot ingest source '%s': path is relative, but repo is not available", source.Name)
		}
		path = filepath.Join(env.RepoDSN, path)
	}

	// get all files in case glob passed
	localPaths, err := doublestar.FilepathGlob(path)
	if err != nil {
		return err
	}
	if len(localPaths) == 0 {
		return fmt.Errorf("file does not exist at %s", conf.Path)
	}

	// Not using query args since not quite sure about behaviour of injecting table names that way.
	// Also, it's a source, so the caller can be trusted.

	from, err := getSourceReader(localPaths, conf)
	if err != nil {
		return err
	}

	qry := fmt.Sprintf("CREATE OR REPLACE TABLE %s AS (SELECT * FROM %s)", source.Name, from)

	return c.Exec(ctx, &drivers.Statement{Query: qry, Priority: 1})
}

func getSourceReader(paths []string, config *connectors.Config) (string, error) {
	format := config.Format
	if format == "" {
		format = fileutil.FullExt(paths[0])
	}

	if format == "" {
		return "", fmt.Errorf("invalid file")
	} else if strings.Contains(format, ".csv") {
		return getCsvSourceReader(paths, config.CSVDelimiter), nil
	} else if strings.Contains(format, ".tsv") || strings.Contains(format, ".txt") {
		return fmt.Sprintf("read_csv_auto(['%s'])", strings.Join(paths, "','")), nil
	} else if strings.Contains(format, ".parquet") {
		return fmt.Sprintf("read_parquet(['%s'])", strings.Join(paths, "','")), nil
	} else {
		return "", fmt.Errorf("file type not supported : %s", format)
	}
}

func getCsvSourceReader(paths []string, delimitter string) string {
	if delimitter == "" {
		return fmt.Sprintf("read_csv_auto(['%s'])", strings.Join(paths, "','"))
	}
	return fmt.Sprintf("read_csv_auto(['%s'], delim='%s')", strings.Join(paths, "','"), delimitter)
}
