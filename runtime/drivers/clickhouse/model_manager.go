package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

type ModelInputProperties struct {
	SQL string `mapstructure:"sql"`
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
	// Columns sets the column names and data types. If unspecified these are detected from the select query by clickhouse.
	// It is also possible to set indexes with this property.
	// Example : (id UInt32, username varchar, email varchar, created_at datetime, INDEX idx1 username TYPE set(100) GRANULARITY 3)
	Columns string `mapstructure:"columns"`
	// Engine sets the table engine. Default: MergeTree
	Engine string `mapstructure:"engine"`
	// OrderBy sets the order by clause. Default: tuple() for MergeTree and not set for other engines
	OrderBy string `mapstructure:"order_by"`
	// PartitionBy sets the partition by clause.
	PartitionBy string `mapstructure:"partition_by"`
	// PrimaryKey sets the primary key clause.
	PrimaryKey string `mapstructure:"primary_key"`
	// SampleBy sets the sample by clause.
	SampleBy string `mapstructure:"sample_by"`
	// TTL sets ttl for column and table.
	TTL string `mapstructure:"ttl"`
	// Settings set the table specific settings.
	Settings string `mapstructure:"settings"`
}

func (p *ModelOutputProperties) Validate(opts *drivers.ModelExecutorOptions) error {
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

	stagingTable, err := olap.InformationSchema().Lookup(ctx, "", "", stagingTableNameFor(res.Table))
	if err == nil {
		_ = olap.DropTable(ctx, stagingTable.Name, stagingTable.View)
	}

	table, err := olap.InformationSchema().Lookup(ctx, "", "", res.Table)
	if err != nil {
		return err
	}

	return olap.DropTable(ctx, table.Name, table.View)
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
		err := olap.RenameTable(ctx, fromName, tmpName, fromIsView)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return olap.RenameTable(ctx, fromName, toName, fromIsView)
}
