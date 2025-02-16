package duckdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type ModelInputProperties struct {
	SQL      string `mapstructure:"sql"`
	Args     []any  `mapstructure:"args"`
	PreExec  string `mapstructure:"pre_exec"`
	PostExec string `mapstructure:"post_exec"`
	// DB is set if sql is to be run against an external database
	DB string `mapstructure:"db"`
}

func (p *ModelInputProperties) Validate() error {
	if p.SQL == "" {
		return fmt.Errorf("missing property 'sql'")
	}
	return nil
}

type ModelOutputProperties struct {
	Table               string                      `mapstructure:"table"`
	Materialize         *bool                       `mapstructure:"materialize"`
	UniqueKey           []string                    `mapstructure:"unique_key"`
	IncrementalStrategy drivers.IncrementalStrategy `mapstructure:"incremental_strategy"`
}

func (p *ModelOutputProperties) Validate(opts *drivers.ModelExecuteOptions) error {
	if opts.Incremental || opts.PartitionRun {
		if p.Materialize != nil && !*p.Materialize {
			return fmt.Errorf("incremental or partitioned models must be materialized")
		}
		p.Materialize = boolPtr(true)
	}

	if opts.InputConnector != opts.OutputConnector {
		if p.Materialize != nil && !*p.Materialize {
			return fmt.Errorf("models that output to a different connector must be materialized")
		}
		p.Materialize = boolPtr(true)
	}

	switch p.IncrementalStrategy {
	case drivers.IncrementalStrategyUnspecified, drivers.IncrementalStrategyAppend, drivers.IncrementalStrategyMerge:
	default:
		return fmt.Errorf("invalid incremental strategy %q", p.IncrementalStrategy)
	}

	if p.IncrementalStrategy == drivers.IncrementalStrategyMerge && len(p.UniqueKey) == 0 {
		return fmt.Errorf(`must specify a "unique_key" when "incremental_strategy" is %q`, p.IncrementalStrategy)
	}

	if p.IncrementalStrategy == drivers.IncrementalStrategyUnspecified {
		if len(p.UniqueKey) == 0 {
			p.IncrementalStrategy = drivers.IncrementalStrategyAppend
		} else {
			p.IncrementalStrategy = drivers.IncrementalStrategyMerge
		}
	}

	return nil
}

type ModelResultProperties struct {
	Table         string `mapstructure:"table"`
	View          bool   `mapstructure:"view"`
	UsedModelName bool   `mapstructure:"used_model_name"`
}

func (c *connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return nil, fmt.Errorf("connector is not an OLAP")
	}

	resProps := &ModelResultProperties{}
	if err := mapstructure.WeakDecode(res.Properties, resProps); err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	if !resProps.UsedModelName {
		return res, nil
	}

	err := olapForceRenameTable(ctx, olap, resProps.Table, resProps.View, newName)
	if err != nil {
		return nil, fmt.Errorf("failed to rename model: %w", err)
	}

	resProps.Table = newName
	resPropsMap := map[string]interface{}{}
	err = mapstructure.WeakDecode(resProps, &resPropsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to encode result properties: %w", err)
	}

	return &drivers.ModelResult{
		Connector:  res.Connector,
		Properties: resPropsMap,
		Table:      newName,
	}, nil
}

func (c *connection) Exists(ctx context.Context, res *drivers.ModelResult) (bool, error) {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return false, fmt.Errorf("connector is not an OLAP")
	}

	_, err := olap.InformationSchema().Lookup(ctx, "", "", res.Table)
	return err == nil, nil
}

func (c *connection) Delete(ctx context.Context, res *drivers.ModelResult) error {
	olap, ok := c.AsOLAP(c.instanceID)
	if !ok {
		return fmt.Errorf("connector is not an OLAP")
	}

	_ = olap.DropTable(ctx, stagingTableNameFor(res.Table))
	return olap.DropTable(ctx, res.Table)
}

func (c *connection) MergePartitionResults(a, b *drivers.ModelResult) (*drivers.ModelResult, error) {
	if a.Table != b.Table {
		return nil, fmt.Errorf("cannot merge partitioned results that output to different table names (table %q is not %q)", a.Table, b.Table)
	}
	return a, nil
}

// stagingTableName returns a stable temporary table name for a destination table.
// By using a stable temporary table name, we can ensure proper garbage collection without managing additional state.
func stagingTableNameFor(table string) string {
	return "__rill_tmp_model_" + table
}

// olapForceRenameTable renames a table or view from fromName to toName in the OLAP connector.
// If a view or table already exists with toName, it is overwritten.
func olapForceRenameTable(ctx context.Context, olap drivers.OLAPStore, fromName string, fromIsView bool, toName string) error {
	if fromName == "" || toName == "" {
		return fmt.Errorf("cannot rename empty table name: fromName=%q, toName=%q", fromName, toName)
	}

	if fromName == toName {
		return nil
	}

	// Infer SQL keyword for the table type
	var typ string
	if fromIsView {
		typ = "VIEW"
	} else {
		typ = "TABLE"
	}

	// Renaming a table to the same name with different casing is not supported. Workaround by renaming to a temporary name first.
	if strings.EqualFold(fromName, toName) {
		tmpName := fmt.Sprintf("__rill_tmp_rename_%s_%s", typ, toName)
		err := olap.RenameTable(ctx, fromName, tmpName)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return olap.RenameTable(ctx, fromName, toName)
}

func boolPtr(b bool) *bool {
	return &b
}
