package rillv1

import (
	"context"
	"fmt"

	"gopkg.in/yaml.v3"
)

var _reservedConnectorNames = map[string]bool{"admin": true, "repo": true, "metastore": true}

// RillYAML is the parsed contents of rill.yaml
type RillYAML struct {
	Title       string
	Description string
	Connectors  []*ConnectorDef
	Variables   []*VariableDef
	Defaults    RillYAMLDefaults
}

// RillYAMLDefaults contains project-wide default YAML properties for different resources
type RillYAMLDefaults struct {
	Sources      yaml.Node
	Models       yaml.Node
	MetricsViews yaml.Node
	Migrations   yaml.Node
}

// ConnectorDef is a subtype of RillYAML, defining connectors required by the project
type ConnectorDef struct {
	Type     string
	Name     string
	Defaults map[string]string
}

// VariableDef is a subtype of RillYAML, defining defaults for project variables
type VariableDef struct {
	Name    string
	Default string
}

// rillYAML is the raw YAML structure of rill.yaml
type rillYAML struct {
	Title       string            `yaml:"title"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env"`
	Connectors  []struct {
		Type     string            `yaml:"type"`
		Name     string            `yaml:"name"`
		Defaults map[string]string `yaml:"defaults"`
	} `yaml:"connectors"`
	Sources    yaml.Node `yaml:"sources"`
	Models     yaml.Node `yaml:"models"`
	Dashboards yaml.Node `yaml:"dashboards"`
	Migrations yaml.Node `yaml:"migrations"`
}

// parseRillYAML parses rill.yaml
func (p *Parser) parseRillYAML(ctx context.Context, path string) error {
	data, err := p.Repo.Get(ctx, path)
	if err != nil {
		return fmt.Errorf("error loading %q: %w", path, err)
	}

	tmp := &rillYAML{}
	if err := yaml.Unmarshal([]byte(data), tmp); err != nil {
		return newYAMLError(err)
	}

	// Validate resource defaults
	if !tmp.Sources.IsZero() {
		if err := tmp.Sources.Decode(&sourceYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Models.IsZero() {
		if err := tmp.Models.Decode(&modelYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Dashboards.IsZero() {
		if err := tmp.Dashboards.Decode(&metricsViewYAML{}); err != nil {
			return newYAMLError(err)
		}
	}
	if !tmp.Migrations.IsZero() {
		if err := tmp.Migrations.Decode(&migrationYAML{}); err != nil {
			return newYAMLError(err)
		}
	}

	res := &RillYAML{
		Title:       tmp.Title,
		Description: tmp.Description,
		Connectors:  make([]*ConnectorDef, len(tmp.Connectors)),
		Variables:   make([]*VariableDef, len(tmp.Env)),
		Defaults: RillYAMLDefaults{
			Sources:      tmp.Sources,
			Models:       tmp.Models,
			MetricsViews: tmp.Dashboards,
			Migrations:   tmp.Migrations,
		},
	}

	for i, c := range tmp.Connectors {
		if _reservedConnectorNames[c.Name] {
			return fmt.Errorf("connector name %q is reserved", c.Name)
		}
		res.Connectors[i] = &ConnectorDef{
			Type:     c.Type,
			Name:     c.Name,
			Defaults: c.Defaults,
		}
	}

	i := 0
	for k, v := range tmp.Env {
		res.Variables[i] = &VariableDef{
			Name:    k,
			Default: v,
		}
		i++
	}

	p.RillYAML = res
	return nil
}
