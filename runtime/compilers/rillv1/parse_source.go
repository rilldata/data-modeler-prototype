package rillv1

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"github.com/robfig/cron/v3"
	"google.golang.org/protobuf/types/known/structpb"

	// Load IANA time zone data
	_ "time/tzdata"
)

// SourceYAML is the raw structure of a Source resource defined in YAML (does not include common fields)
type SourceYAML struct {
	commonYAML `yaml:",inline" mapstructure:",squash"` // Only to avoid loading common fields into Properties
	Timeout    string                                  `yaml:"timeout"`
	Refresh    *ScheduleYAML                           `yaml:"refresh"`
	Properties map[string]any                          `yaml:",inline" mapstructure:",remain"`
}

// parseSource parses a source definition and adds the resulting resource to p.Resources.
func (p *Parser) parseSource(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &SourceYAML{}
	err := p.decodeNodeYAML(node, false, tmp)
	if err != nil {
		return err
	}

	// Backwards compatibility: "type:" was previously used instead of "connector:".
	// So if "type:" is not a valid resource kind, we treat it as a connector.
	if tmp.Type != nil {
		if _, err := ParseResourceKind(*tmp.Type); err != nil {
			node.Connector = *tmp.Type
			node.ConnectorInferred = false
		}
	}

	// If the source has SQL and hasn't specified a connector, we treat it as a model
	if node.SQL != "" && node.ConnectorInferred {
		return p.parseModel(ctx, node)
	}

	// Add SQL as a property
	if node.SQL != "" {
		if tmp.Properties == nil {
			tmp.Properties = map[string]any{}
		}
		tmp.Properties["sql"] = strings.TrimSpace(node.SQL)
	}

	// Parse timeout
	var timeout time.Duration
	if tmp.Timeout != "" {
		timeout, err = parseDuration(tmp.Timeout)
		if err != nil {
			return err
		}
	}

	// Parse refresh schedule
	schedule, err := p.parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	// Backward compatibility: when the default connector is "olap", and it's a DuckDB connector, a source with connector "duckdb" should run on it
	if p.DefaultOLAPConnector == "olap" && node.Connector == "duckdb" {
		node.Connector = "olap"
	}

	// Validate the source has a connector
	if node.ConnectorInferred {
		return fmt.Errorf("must explicitly specify a connector for sources")
	}

	props, err := structpb.NewStruct(tmp.Properties)
	if err != nil {
		return fmt.Errorf("encountered invalid property type: %w", err)
	}

	outputProps, err := structpb.NewStruct(map[string]any{"materialize": true})
	if err != nil {
		return fmt.Errorf("internal error: can not create output properties for source: %w", err)
	}

	// Track as a model
	// We allowed a special resource type (source) to ingest data from external connector
	// After the unification of sources and models everything is a model
	r, err := p.insertResource(ResourceKindModel, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	if schedule != nil {
		r.ModelSpec.RefreshSchedule = schedule
	}

	if timeout > 0 {
		r.ModelSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	r.ModelSpec.InputConnector = node.Connector
	r.ModelSpec.InputProperties = mergeStructPB(r.ModelSpec.InputProperties, props)

	r.ModelSpec.OutputConnector = p.defaultOLAPConnector()
	r.ModelSpec.OutputProperties = outputProps
	return nil
}

// ScheduleYAML is the raw structure of a refresh schedule clause defined in YAML.
// This does not represent a stand-alone YAML file, just a partial used in other structs.
type ScheduleYAML struct {
	RefUpdate *bool  `yaml:"ref_update" mapstructure:"ref_update"`
	Cron      string `yaml:"cron" mapstructure:"cron"`
	Every     string `yaml:"every" mapstructure:"every"`
	TimeZone  string `yaml:"time_zone" mapstructure:"time_zone"`
	Disable   bool   `yaml:"disable" mapstructure:"disable"`
	RunInDev  bool   `yaml:"run_in_dev" mapstructure:"run_in_dev"`
}

func (p *Parser) parseScheduleYAML(raw *ScheduleYAML) (*runtimev1.Schedule, error) {
	s := &runtimev1.Schedule{
		RefUpdate: true, // By default, refresh on updates to refs
	}

	if raw == nil {
		return s, nil
	}

	if raw.Disable {
		s.RefUpdate = false
		s.Disable = true
		return s, nil
	}

	// Enforce run_in_dev only for scheduled refreshes. We always honor ref_update even in dev.
	skipScheduledRefresh := !raw.RunInDev && p.isDev()

	if raw.RefUpdate != nil {
		s.RefUpdate = *raw.RefUpdate
	}

	if !skipScheduledRefresh && raw.Cron != "" {
		_, err := cron.ParseStandard(raw.Cron)
		if err != nil {
			return nil, fmt.Errorf("invalid cron schedule: %w", err)
		}
		s.Cron = raw.Cron
	}

	if !skipScheduledRefresh && raw.Every != "" {
		d, err := parseDuration(raw.Every)
		if err != nil {
			return nil, fmt.Errorf("invalid ticker: %w", err)
		}
		s.TickerSeconds = uint32(d.Seconds())
	}

	if raw.TimeZone != "" {
		_, err := time.LoadLocation(raw.TimeZone)
		if err != nil {
			return nil, fmt.Errorf("invalid time zone: %w", err)
		}
		s.TimeZone = raw.TimeZone
	}

	return s, nil
}

// parseDuration parses a value into a time duration.
// If no unit is specified, it assumes the value is in seconds.
func parseDuration(v any) (time.Duration, error) {
	switch v := v.(type) {
	case int:
		return time.Duration(v) * time.Second, nil
	case string:
		// Try parsing as an int first
		res, err := strconv.Atoi(v)
		if err == nil {
			return time.Duration(res) * time.Second, nil
		}
		// Try parsing as a Go duration string
		d, err := time.ParseDuration(v)
		if err == nil {
			return d, nil
		}
		// Try parsing as an ISO 8601 duration string
		id, err := duration.ParseISO8601(v)
		if err == nil {
			d, ok := id.EstimateNative()
			if !ok {
				return 0, fmt.Errorf("time duration string %q can't be resolved to an absolute duration", v)
			}
			return d, nil
		}
		// Give up
		return 0, fmt.Errorf("invalid time duration string %q", v)
	default:
		return 0, fmt.Errorf("invalid time duration value <%v>", v)
	}
}

// mergeStructPB merges two structpb.Structs, with b taking precedence over a.
// It overwrites a in place and returns it.
func mergeStructPB(a, b *structpb.Struct) *structpb.Struct {
	if a == nil || a.Fields == nil {
		return b
	}
	if b == nil || b.Fields == nil {
		return a
	}
	for k, v := range b.Fields {
		a.Fields[k] = v
	}
	return a
}
