package duckdb

import (
	"context"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

type localFileToSelfExecutor struct {
	c    *connection
	from drivers.FileStore
}

var _ drivers.ModelExecutor = &localFileToSelfExecutor{}

type inputProps struct {
	InvalidateOnChange bool           `mapstructure:"invalidate_on_change"`
	Format             string         `mapstructure:"format"`
	DuckDB             map[string]any `mapstructure:"duckdb"`
}

func (p *inputProps) Validate() error {
	return nil
}

func (e *localFileToSelfExecutor) Concurrency(desired int) (int, bool) {
	if desired > 1 {
		return 0, false
	}
	return 1, true
}

func (e *localFileToSelfExecutor) Execute(ctx context.Context, opts *drivers.ModelExecuteOptions) (*drivers.ModelResult, error) {
	inputProps := &inputProps{}
	if err := mapstructure.WeakDecode(opts.InputProperties, inputProps); err != nil {
		return nil, fmt.Errorf("failed to parse input properties: %w", err)
	}
	if err := inputProps.Validate(); err != nil {
		return nil, fmt.Errorf("invalid input properties: %w", err)
	}

	outputProps := &ModelOutputProperties{}
	if err := mapstructure.WeakDecode(opts.OutputProperties, outputProps); err != nil {
		return nil, fmt.Errorf("failed to parse output properties: %w", err)
	}
	if err := outputProps.Validate(opts); err != nil {
		return nil, fmt.Errorf("invalid output properties: %w", err)
	}

	usedModelName := false
	if outputProps.Table == "" {
		outputProps.Table = opts.ModelName
		usedModelName = true
	}

	materialize := true
	if outputProps.Materialize != nil {
		materialize = *outputProps.Materialize
	}

	asView := !materialize
	tableName := outputProps.Table

	// Prepare for ingesting into the staging view/table.
	// NOTE: This intentionally drops the end table if not staging changes.
	stagingTableName := tableName
	if opts.Env.StageChanges {
		stagingTableName = stagingTableNameFor(tableName)
	}
	_ = e.c.DropTable(ctx, stagingTableName)

	// get the local file path
	localPaths, err := e.from.FilePaths(ctx, opts.InputProperties)
	if err != nil {
		return nil, err
	}
	if len(localPaths) == 0 {
		return nil, fmt.Errorf("no files to ingest")
	}

	if inputProps.Format == "" {
		inputProps.Format = fileutil.FullExt(localPaths[0])
	} else {
		inputProps.Format = "." + inputProps.Format
	}

	from, err := sourceReader(localPaths, inputProps.Format, inputProps.DuckDB)
	if err != nil {
		return nil, err
	}

	// create the table
	err = e.c.CreateTableAsSelect(ctx, stagingTableName, "SELECT * FROM "+from, &drivers.CreateTableOptions{View: asView})
	if err != nil {
		_ = e.c.DropTable(ctx, stagingTableName)
		return nil, fmt.Errorf("failed to create model: %w", err)
	}

	// Rename the staging table to the final table name
	if stagingTableName != tableName {
		err = olapForceRenameTable(ctx, e.c, stagingTableName, asView, tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to rename staged model: %w", err)
		}
	}

	// Build result props
	resultProps := &ModelResultProperties{
		Table:         tableName,
		View:          asView,
		UsedModelName: usedModelName,
	}
	resultPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resultProps, &resultPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	// Done
	return &drivers.ModelResult{
		Connector:  opts.OutputConnector,
		Properties: resultPropsMap,
		Table:      tableName,
	}, nil
}
