package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
)

const _defaultConcurrentInserts = 1

type ModelInputProperties struct {
	SQL string `mapstructure:"sql"`
}

func (p *ModelInputProperties) Validate() error {
	return nil
}

type ModelOutputProperties struct {
	Table               string                      `mapstructure:"table"`
	Materialize         *bool                       `mapstructure:"materialize"`
	UniqueKey           []string                    `mapstructure:"unique_key"`
	IncrementalStrategy drivers.IncrementalStrategy `mapstructure:"incremental_strategy"`
	// Typ to materialize the model into. Possible values include `TABLE`, `VIEW` or `DICTIONARY`. Optional.
	Typ string `mapstructure:"type"`
	// Columns sets the column names and data types. If unspecified these are detected from the select query by clickhouse.
	// It is also possible to set indexes with this property.
	// Example : (id UInt32, username varchar, email varchar, created_at datetime, INDEX idx1 username TYPE set(100) GRANULARITY 3)
	Columns string `mapstructure:"columns"`
	// Config can be used to set the table parameters like engine, partition key in SQL format without setting individual properties.
	// It also allows creating dictionaries using a source.
	// Example:
	//  ENGINE = MergeTree
	//	PARTITION BY toYYYYMM(__time)
	//	ORDER BY __time
	//	TTL d + INTERVAL 1 MONTH DELETE
	Config string `mapstructure:"config"`
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
	// TableSettings set the table specific settings.
	TableSettings string `mapstructure:"table_settings"`
	// QuerySettings sets the settings clause used in insert/create table as select queries.
	QuerySettings string `mapstructure:"query_settings"`
	// OnCluster adds a ON CLUSTER clause to the create table statement. Optional.
	// The `cluster` property must be set in the connector configuration.
	OnCluster bool `mapstructure:"on_cluster"`
	// DistributedConfig is config for distributed table.
	// Note: the table name in config should be table__local. Optional.
	// TODO :: How to handle staged changes ?
	DistributedConfig string `mapstructure:"distributed_config"`
}

func (p *ModelOutputProperties) Validate(opts *drivers.ModelExecuteOptions) error {
	if p.Config != "" {
		if p.Engine != "" || p.OrderBy != "" || p.PartitionBy != "" || p.PrimaryKey != "" || p.SampleBy != "" || p.TTL != "" || p.TableSettings != "" {
			return fmt.Errorf("`config` property cannot be used with individual properties")
		}
	}
	p.Typ = strings.ToUpper(p.Typ)
	if p.Typ != "" && p.Materialize != nil {
		return fmt.Errorf("cannot set both `type` and `materialize` properties")
	}
	if p.Materialize != nil {
		if *p.Materialize {
			p.Typ = "TABLE"
		} else {
			p.Typ = "VIEW"
		}
	}
	if opts.Incremental || opts.SplitRun {
		if p.Typ != "" && p.Typ != "TABLE" {
			return fmt.Errorf("incremental or split models must be materialized")
		}
		p.Typ = "TABLE"
	}
	if p.Typ == "" {
		p.Typ = "VIEW"
	}

	if p.Typ == "DICTIONARY" && p.Columns == "" {
		return fmt.Errorf("model materialized as dictionary must specify columns")
	}

	switch p.IncrementalStrategy {
	case drivers.IncrementalStrategyUnspecified, drivers.IncrementalStrategyAppend:
	default:
		return fmt.Errorf("invalid incremental strategy %q", p.IncrementalStrategy)
	}

	if p.IncrementalStrategy == drivers.IncrementalStrategyUnspecified {
		p.IncrementalStrategy = drivers.IncrementalStrategyAppend
	}
	return nil
}

func (p *ModelOutputProperties) tblConfig() string {
	if p.Config != "" {
		return p.Config
	}
	var sb strings.Builder
	// engine with default
	if p.Engine != "" {
		fmt.Fprintf(&sb, "ENGINE = %s", p.Engine)
	} else {
		fmt.Fprintf(&sb, "ENGINE = MergeTree")
	}

	// order_by
	if p.OrderBy != "" {
		fmt.Fprintf(&sb, " ORDER BY %s", p.OrderBy)
	} else if p.Engine == "MergeTree" {
		// need ORDER BY for MergeTree
		// it is optional for many other engines
		fmt.Fprintf(&sb, " ORDER BY tuple()")
	}

	// partition_by
	if p.PartitionBy != "" {
		fmt.Fprintf(&sb, " PARTITION BY %s", p.PartitionBy)
	}

	// primary_key
	if p.PrimaryKey != "" {
		fmt.Fprintf(&sb, " PRIMARY KEY %s", p.PrimaryKey)
	}

	// sample_by
	if p.SampleBy != "" {
		fmt.Fprintf(&sb, " SAMPLE BY %s", p.SampleBy)
	}

	// ttl
	if p.TTL != "" {
		fmt.Fprintf(&sb, " TTL %s", p.TTL)
	}

	// settings
	if p.TableSettings != "" {
		fmt.Fprintf(&sb, " %s", p.TableSettings)
	}
	return sb.String()
}

type ModelResultProperties struct {
	Table         string `mapstructure:"table"`
	View          bool   `mapstructure:"view"`
	UsedModelName bool   `mapstructure:"used_model_name"`
}

func (c *connection) Rename(ctx context.Context, res *drivers.ModelResult, newName string, env *drivers.ModelEnv) (*drivers.ModelResult, error) {
	resProps := &ModelResultProperties{}
	if err := mapstructure.WeakDecode(res.Properties, resProps); err != nil {
		return nil, fmt.Errorf("failed to parse previous result properties: %w", err)
	}

	if !resProps.UsedModelName {
		return res, nil
	}

	err := olapForceRenameTable(ctx, c, resProps.Table, resProps.View, newName)
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
		_ = c.DropTable(ctx, stagingTable.Name, stagingTable.View)
	}

	table, err := olap.InformationSchema().Lookup(ctx, "", "", res.Table)
	if err != nil {
		return err
	}

	return c.DropTable(ctx, table.Name, table.View)
}

func (c *connection) MergeSplitResults(a, b *drivers.ModelResult) (*drivers.ModelResult, error) {
	if a.Table != b.Table {
		return nil, fmt.Errorf("cannot merge split results that output to different table names (%q != %q)", a.Table, b.Table)
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
func olapForceRenameTable(ctx context.Context, c *connection, fromName string, fromIsView bool, toName string) error {
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
		err := c.RenameTable(ctx, fromName, tmpName, fromIsView)
		if err != nil {
			return err
		}
		fromName = tmpName
	}

	// Do the rename
	return c.RenameTable(ctx, fromName, toName, fromIsView)
}

func boolPtr(b bool) *bool {
	return &b
}
