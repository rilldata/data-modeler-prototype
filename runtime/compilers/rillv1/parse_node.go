package rillv1

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/pkg/sqlparse"
	"gopkg.in/yaml.v3"
)

// Node represents one path stem in the project. It contains data derived from a YAML and/or SQL file (e.g. "/path/to/file.yaml" for "/path/to/file.sql").
type Node struct {
	Version           int
	Kind              ResourceKind
	Name              string
	Refs              []ResourceName
	Paths             []string
	YAML              *yaml.Node
	YAMLOverride      *yaml.Node
	YAMLRaw           string
	YAMLPath          string
	Connector         string
	ConnectorInferred bool
	SQL               string
	SQLPath           string
	SQLAnnotations    map[string]any
	SQLUsesTemplating bool
}

// parseNode multiplexes to the appropriate parse function based on the node kind.
func (p *Parser) parseNode(ctx context.Context, node *Node) error {
	switch node.Kind {
	case ResourceKindSource:
		return p.parseSource(ctx, node)
	case ResourceKindModel:
		return p.parseModel(ctx, node)
	case ResourceKindMetricsView:
		return p.parseMetricsView(node)
	case ResourceKindExplore:
		return p.parseExplore(node)
	case ResourceKindMigration:
		return p.parseMigration(node)
	case ResourceKindReport:
		return p.parseReport(node)
	case ResourceKindAlert:
		return p.parseAlert(node)
	case ResourceKindTheme:
		return p.parseTheme(node)
	case ResourceKindComponent:
		return p.parseComponent(node)
	case ResourceKindCanvas:
		return p.parseCanvas(node)
	case ResourceKindAPI:
		return p.parseAPI(node)
	case ResourceKindConnector:
		return p.parseConnector(node)
	default:
		panic(fmt.Errorf("unexpected resource type: %s", node.Kind.String()))
	}
}

// commonYAML parses YAML fields common to all YAML files.
type commonYAML struct {
	// Version of the parser to use for this file. Enables backwards compatibility for breaking changes.
	Version int `yaml:"version"`
	// Type can be inferred from the directory name in certain cases, but otherwise must be specified manually.
	Type *string `yaml:"type"`
	// Deprecated: Changed to Type. "Kind" is still used internally to refer to resource types.
	Kind *string `yaml:"kind"`
	// Name is usually inferred from the filename, but can be specified manually.
	Name string `yaml:"name"`
	// Namespace is an optional value to group resources by.
	// It currently just gets pre-pended to the resource name in the format `<namespace>/<name>`.
	Namespace string `yaml:"namespace"`
	// Refs are a list of other resources that this resource depends on. They are usually inferred from other fields, but can also be specified manually.
	Refs []yaml.Node `yaml:"refs"`
	// ParserConfig enables setting file-level parser config.
	ParserConfig struct {
		Templating *bool `yaml:"templating"`
	} `yaml:"parser"`
	// Connector names the connector to use for this resource. It may not be used in all resources, but is included here since it provides context for the SQL field.
	Connector string `yaml:"connector"`
	// SQL contains the SQL string for this resource. It may be specified inline, or will be loaded from a file at the same stem. It may not be supported in all resources.
	SQL string `yaml:"sql"`
	// Environment-specific overrides
	EnvironmentOverrides map[string]yaml.Node `yaml:"environment_overrides"`
	// Deprecated key for environment-specific overrides (replaced by "environment_overrides")
	EnvironmentOverridesOld map[string]yaml.Node `yaml:"env"`
	// Shorthand for setting "environment_overrides:dev:"
	Dev yaml.Node `yaml:"dev"`
	// Shorthand for setting "environment_overrides:prod:"
	Prod yaml.Node `yaml:"prod"`
}

// parseStem parses a pair of YAML and SQL files with the same path stem (e.g. "/path/to/file.yaml" for "/path/to/file.sql").
// Note that either of the YAML or SQL files may be empty (the paths arg will only contain non-nil paths).
func (p *Parser) parseStem(paths []string, ymlPath, yml, sqlPath, sql string) (*Node, error) {
	// The rest of the function builds a Node from YAML and SQL info
	res := &Node{Paths: paths}

	// Parse YAML into commonYAML
	var cfg *commonYAML
	if ymlPath != "" {
		var node yaml.Node
		err := yaml.Unmarshal([]byte(yml), &node)
		if err != nil {
			return nil, pathError{path: ymlPath, err: newYAMLError(err)}
		}
		res.YAML = &node
		res.YAMLRaw = yml
		res.YAMLPath = ymlPath

		err = node.Decode(&cfg)
		if err != nil {
			return nil, pathError{path: ymlPath, err: newYAMLError(err)}
		}
	}

	// Handle YAML config
	templatingEnabled := true
	if cfg != nil {
		// Copy EnvironmentOverridesOld to EnvironmentOverrides
		for k, v := range cfg.EnvironmentOverridesOld { // nolint: gocritic // Using a pointer changes parser behavior
			if cfg.EnvironmentOverrides == nil {
				cfg.EnvironmentOverrides = make(map[string]yaml.Node)
			}
			cfg.EnvironmentOverrides[k] = v
		}

		// Handle "dev:" and "prod:" shorthands (copy to to cfg.Env)
		if !cfg.Dev.IsZero() {
			if cfg.EnvironmentOverrides == nil {
				cfg.EnvironmentOverrides = make(map[string]yaml.Node)
			}
			cfg.EnvironmentOverrides["dev"] = cfg.Dev
		}
		if !cfg.Prod.IsZero() {
			if cfg.EnvironmentOverrides == nil {
				cfg.EnvironmentOverrides = make(map[string]yaml.Node)
			}
			cfg.EnvironmentOverrides["prod"] = cfg.Prod
		}

		// Set environment-specific override
		if envOverride := cfg.EnvironmentOverrides[p.Environment]; !envOverride.IsZero() {
			res.YAMLOverride = &envOverride

			// Apply the override immediately in case it changes any of the commonYAML fields
			err := res.YAMLOverride.Decode(&cfg)
			if err != nil {
				return nil, pathError{path: ymlPath, err: newYAMLError(err)}
			}
		}

		// Copy basic properties
		res.Version = cfg.Version
		res.Name = cfg.Name
		res.Connector = cfg.Connector
		res.SQL = cfg.SQL
		res.SQLPath = ymlPath

		// Handle templating config
		if cfg.ParserConfig.Templating != nil {
			templatingEnabled = *cfg.ParserConfig.Templating
		}

		// Parse refs provided in YAML
		var err error
		res.Refs, err = parseYAMLRefs(cfg.Refs)
		if err != nil {
			return nil, pathError{path: ymlPath, err: newYAMLError(err)}
		}

		// Parse resource kind if set in YAML. If not set, we try to infer it from path later.
		// NOTE: We use "kind" internally, but "type:" is the preferred user-facing field.
		// However, we still need to support "kind:" for backwards compatibility.
		if cfg.Kind != nil {
			res.Kind, err = ParseResourceKind(*cfg.Kind)
			if err != nil {
				return nil, pathError{path: ymlPath, err: err}
			}
		}
		if cfg.Type != nil {
			kind, err := ParseResourceKind(*cfg.Type)
			if err == nil {
				res.Kind = kind
			} else if !strings.HasPrefix(paths[0], "/sources") {
				// Backwards compatibility: "type:" was previously used in sources instead of "connector:". This was when sources were always created in the "sources" directory.
				// So for backwards compatibility, we ignore parse errors for the "type:" field when the file is in the "sources" directory.
				// (The source parser handles the backwards compatibility around mapping "type:" to "connector:".)
				return nil, pathError{path: ymlPath, err: err}
			}
		}
	}

	// Set SQL
	if sql != "" {
		// Check SQL was not already provided in YAML
		if res.SQL != "" {
			return nil, pathError{path: ymlPath, err: errors.New("SQL provided using both a YAML key and a companion file")}
		}
		res.SQL = sql
		res.SQLPath = sqlPath
	}

	// Parse SQL templating
	if templatingEnabled && res.SQL != "" {
		meta, err := AnalyzeTemplate(res.SQL)
		if err != nil {
			if sqlPath != "" {
				return nil, pathError{path: sqlPath, err: err}
			}
			return nil, pathError{path: ymlPath, err: err}
		}

		res.SQLUsesTemplating = meta.UsesTemplating
		res.SQLAnnotations = meta.Config
		res.Refs = append(res.Refs, meta.Refs...) // If needed, deduplication happens in insertResource

		// Additionally parse annotations provided in comments (e.g. "-- @materialize: true")
		commentAnnotations := sqlparse.ExtractAnnotations(res.SQL)
		if len(commentAnnotations) > 0 && res.SQLAnnotations == nil {
			res.SQLAnnotations = make(map[string]any)
		}
		for k, v := range commentAnnotations {
			res.SQLAnnotations[k] = v
		}

		// Expand dots in annotations. E.g. turn annotations["foo.bar"] into annotations["foo"]["bar"].
		res.SQLAnnotations, err = expandAnnotations(res.SQLAnnotations)
		if err != nil {
			if sqlPath != "" {
				return nil, pathError{path: sqlPath, err: err}
			}
			return nil, pathError{path: ymlPath, err: err}
		}
	}

	// Some annotations in the SQL file can override the base config: kind, name, connector
	var err error
	for k, v := range res.SQLAnnotations {
		switch strings.ToLower(k) {
		case "type", "kind": // "kind" is for backwards compatibility
			v, ok := v.(string)
			if !ok {
				err = fmt.Errorf("invalid type %T for property 'type'", v)
				break
			}
			res.Kind, err = ParseResourceKind(v)
			if err != nil {
				break
			}
		case "name":
			v, ok := v.(string)
			if !ok {
				err = fmt.Errorf("invalid type %T for property 'name'", v)
				break
			}
			res.Name = v
		case "connector":
			v, ok := v.(string)
			if !ok {
				err = fmt.Errorf("invalid type %T for property 'connector'", v)
				break
			}
			res.Connector = v
		}
	}
	if err != nil {
		if sqlPath != "" {
			return nil, pathError{path: sqlPath, err: err}
		}
		return nil, pathError{path: ymlPath, err: err}
	}

	// If name is not set in YAML or SQL, infer it from path
	if res.Name == "" {
		if ymlPath != "" {
			res.Name = filepath.Base(pathStem(ymlPath))
		} else if sqlPath != "" {
			res.Name = filepath.Base(pathStem(sqlPath))
		}
	}

	// If a namespace was provided in YAML, prepend it to the name.
	if cfg != nil && cfg.Namespace != "" {
		res.Name = cfg.Namespace + ":" + res.Name
	}

	// If resource kind is not set in YAML or SQL, try to infer it from the context
	if res.Kind == ResourceKindUnspecified {
		if strings.HasPrefix(paths[0], "/sources") {
			res.Kind = ResourceKindSource
		} else if strings.HasPrefix(paths[0], "/models") {
			res.Kind = ResourceKindModel
		} else if strings.HasPrefix(paths[0], "/dashboards") {
			res.Kind = ResourceKindMetricsView
		} else if strings.HasPrefix(paths[0], "/connectors") {
			res.Kind = ResourceKindConnector
		} else if strings.HasPrefix(paths[0], "/init.sql") {
			res.Kind = ResourceKindMigration
		} else if sqlPath != "" {
			// SQL files without an explicit kind are assumed to be models
			res.Kind = ResourceKindModel
		} else {
			path := ymlPath
			if path == "" {
				path = sqlPath
			}
			return nil, pathError{path: path, err: errors.New("resource type not specified and could not be inferred from context")}
		}
	}

	// If connector wasn't set explicitly, use the default OLAP connector
	if res.Connector == "" {
		res.Connector = p.defaultOLAPConnector()
		res.ConnectorInferred = true
	}

	return res, nil
}

// decodeNodeYAML decodes a Node into a YAML struct.
// If knownFields is true, it will return an error if the YAML contains unknown fields.
// It applies defaults from rill.yaml, then the YAML, then the YAML's environment-specific overrides, and finally the SQL annotations.
// If an error is returned, it will be a pathError associated with the node.
func (p *Parser) decodeNodeYAML(node *Node, knownFields bool, dst any) error {
	// Apply defaults from rill.yaml
	if p.RillYAML != nil {
		defaults := p.RillYAML.Defaults[node.Kind]
		if !defaults.IsZero() {
			if err := defaults.Decode(dst); err != nil {
				return pathError{path: node.YAMLPath, err: fmt.Errorf("failed applying defaults from rill.yaml: %w", err)}
			}
		}
	}

	// Apply YAML
	if node.YAML != nil {
		var err error
		if knownFields {
			// Using node.YAMLRaw instead of node.YAML because we need to set KnownFields for metrics views
			dec := yaml.NewDecoder(strings.NewReader(node.YAMLRaw))
			dec.KnownFields(true)
			err = dec.Decode(dst)
		} else {
			err = node.YAML.Decode(dst)
		}
		if err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	// Override YAML config with SQL annotations
	if len(node.SQLAnnotations) > 0 {
		err := mapstructureUnmarshal(node.SQLAnnotations, dst)
		if err != nil {
			return pathError{path: node.SQLPath, err: fmt.Errorf("invalid SQL annotations: %w", err)}
		}
	}

	// Apply environment-specific overrides
	if node.YAMLOverride != nil {
		err := node.YAMLOverride.Decode(dst)
		if err != nil {
			return pathError{path: node.YAMLPath, err: newYAMLError(err)}
		}
	}

	return nil
}

// parseYAMLRefs parses a list of YAML nodes into a list of ResourceNames.
// It's used to parse the "refs" field in baseConfig.
func parseYAMLRefs(refs []yaml.Node) ([]ResourceName, error) {
	var res []ResourceName
	for i := range refs {
		ref := refs[i]

		// We support string refs of the form "my-resource" and "Kind/my-resource"
		if ref.Kind == yaml.ScalarNode {
			var identifier string
			err := ref.Decode(&identifier)
			if err != nil {
				return nil, fmt.Errorf("invalid refs: %v", ref)
			}

			// Parse name and kind from identifier
			parts := strings.Split(identifier, "/")
			if len(parts) != 1 && len(parts) != 2 {
				return nil, fmt.Errorf("invalid refs: invalid identifier %q", identifier)
			}

			var name ResourceName
			if len(parts) == 1 {
				name.Name = parts[0]
			} else {
				// Kind and name specified
				kind, err := ParseResourceKind(parts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid refs: %w", err)
				}
				name.Kind = kind
				name.Name = parts[1]
			}
			res = append(res, name)
			continue
		}

		// We support map refs of the form { type: "kind", name: "my-resource" }
		if ref.Kind == yaml.MappingNode {
			m := make(map[string]string)
			err := ref.Decode(m)
			if err != nil {
				return nil, fmt.Errorf("invalid refs: %w", err)
			}
			if m["name"] == "" {
				return nil, errors.New(`an object ref must provide the properties "type" and "name" properties`)
			}

			var name ResourceName
			name.Name = m["name"]

			if m["type"] != "" {
				kind, err := ParseResourceKind(m["type"])
				if err != nil {
					return nil, fmt.Errorf("invalid refs: %w", err)
				}
				name.Kind = kind
			}

			// Backwards compatibility for "kind:" changed to "type:"
			if m["kind"] != "" {
				kind, err := ParseResourceKind(m["kind"])
				if err != nil {
					return nil, fmt.Errorf("invalid refs: %w", err)
				}
				name.Kind = kind
			}

			res = append(res, name)
			continue
		}

		// ref was neither a string nor a map
		return nil, fmt.Errorf("invalid refs: %v", ref)
	}
	for i := range res {
		// Source is deprecated but for backwards compatibility, we convert it to model
		if res[i].Kind == ResourceKindSource {
			res[i].Kind = ResourceKindModel
		}
	}
	return res, nil
}

// mapstructureUnmarshal is used to unmarshal SQL annotations into a struct (overriding YAML config).
func mapstructureUnmarshal(annotations map[string]any, dst any) error {
	if len(annotations) == 0 {
		return nil
	}
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           dst,
		WeaklyTypedInput: true,
	})
	if err != nil {
		panic(err)
	}
	return decoder.Decode(annotations)
}

// expandAnnotations turns annotations with dots in their key into nested maps.
// For example, annotations["foo.bar"] becomes annotations["foo"]["bar"].
func expandAnnotations(annotations map[string]any) (map[string]any, error) {
	if len(annotations) == 0 {
		return nil, nil
	}
	res := make(map[string]any)
	for k, v := range annotations {
		parts := strings.Split(k, ".")
		if len(parts) < 2 {
			res[k] = v
			continue
		}

		m := res
		for i := 0; i < len(parts)-1; i++ {
			part := parts[i]

			// Check if a map already exists for this part; if yes, assign to m
			x, ok := m[part]
			if ok {
				m, ok = x.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("invalid annotation %q: nesting incompatible with other keys", k)
				}
				continue
			}

			// Create a map for this part, then update m
			tmp := make(map[string]any)
			m[part] = tmp
			m = tmp
		}

		// Check the last part of this key isn't an intermediate part of a previously expanded key
		k2 := parts[len(parts)-1]
		if _, ok := m[k2]; ok {
			return nil, fmt.Errorf("invalid annotation2 %q: nesting incompatible with other keys", k)
		}

		// Assign the value to the last part
		m[k2] = v
	}
	return res, nil
}
