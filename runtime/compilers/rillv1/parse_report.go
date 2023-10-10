package rillv1

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/pkg/duration"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
)

// reportYAML is the raw structure of a Report resource defined in YAML (does not include common fields)
type reportYAML struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Disabled   bool             `yaml:"disabled"`
	Refresh    *scheduleYAML    `yaml:"refresh"`
	Timeout    string           `yaml:"timeout"`
	Operation  struct {
		Name       string         `yaml:"name"`
		TimeRange  string         `yaml:"time_range"`
		Properties map[string]any `yaml:"properties"`
	} `yaml:"operation"`
	Export struct {
		Format string `yaml:"format"`
		Limit  uint   `yaml:"limit"`
	} `yaml:"export"`
	Recipients []string `yaml:"recipients"`
	Template   struct {
		OpenURL string `yaml:"open_url"`
		EditURL string `yaml:"edit_url"`
	} `yaml:"template"`
}

// parseReport parses a report definition and adds the resulting resource to p.Resources.
func (p *Parser) parseReport(ctx context.Context, node *Node) error {
	// Parse YAML
	tmp := &reportYAML{}
	if node.YAMLRaw != "" {
		// Can't use node.YAML because we want to set KnownFields for reports
		dec := yaml.NewDecoder(strings.NewReader(node.YAMLRaw))
		dec.KnownFields(true)
		if err := dec.Decode(tmp); err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("reports cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("reports cannot have a connector")
	}

	// Parse refresh schedule
	schedule, err := parseScheduleYAML(tmp.Refresh)
	if err != nil {
		return err
	}

	// Parse timeout
	var timeout time.Duration
	if tmp.Timeout != "" {
		timeout, err = parseDuration(tmp.Timeout)
		if err != nil {
			return err
		}
	}

	// Validate and parse operation params
	if tmp.Operation.Name == "" {
		return fmt.Errorf(`invalid value %q for property "operation.name"`, tmp.Operation.Name)
	}
	operationProps, err := structpb.NewStruct(tmp.Operation.Properties)
	if err != nil {
		return fmt.Errorf("failed to serialize properties: %w", err)
	}
	if tmp.Operation.TimeRange != "" {
		_, err := duration.ParseISO8601(tmp.Operation.TimeRange)
		if err != nil {
			return fmt.Errorf(`invalid "operation.dynamic_time_range": %w`, err)
		}
	}

	// Parse export format
	exportFormat, err := parseExportFormat(tmp.Export.Format)
	if err != nil {
		return err
	}
	if exportFormat == runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED {
		return fmt.Errorf(`missing required property "export.format"`)
	}

	// Validate recipients
	if len(tmp.Recipients) == 0 {
		return fmt.Errorf(`missing required property "recipients"`)
	}
	for _, email := range tmp.Recipients {
		_, err := mail.ParseAddress(email)
		if err != nil {
			return fmt.Errorf("invalid recipient email address %q", email)
		}
	}

	// Track report
	r, err := p.insertResource(ResourceKindReport, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ReportSpec.Title = tmp.Title
	r.ReportSpec.Disabled = tmp.Disabled
	if schedule != nil {
		r.ReportSpec.RefreshSchedule = schedule
	}
	if timeout != 0 {
		r.SourceSpec.TimeoutSeconds = uint32(timeout.Seconds())
	}
	r.ReportSpec.OperationName = tmp.Operation.Name
	r.ReportSpec.OperationProperties = operationProps
	r.ReportSpec.OperationTimeRange = tmp.Operation.TimeRange
	r.ReportSpec.ExportLimit = uint32(tmp.Export.Limit)
	r.ReportSpec.ExportFormat = exportFormat
	r.ReportSpec.Recipients = tmp.Recipients

	return nil
}

func parseExportFormat(s string) (runtimev1.ExportFormat, error) {
	switch strings.ToLower(s) {
	case "":
		return runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED, nil
	case "csv":
		return runtimev1.ExportFormat_EXPORT_FORMAT_CSV, nil
	case "xlsx":
		return runtimev1.ExportFormat_EXPORT_FORMAT_XLSX, nil
	case "parquet":
		return runtimev1.ExportFormat_EXPORT_FORMAT_PARQUET, nil
	default:
		return runtimev1.ExportFormat_EXPORT_FORMAT_UNSPECIFIED, fmt.Errorf("invalid export format %q", s)
	}
}
