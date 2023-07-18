package rillv1

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/sqlparse"
	"gopkg.in/yaml.v3"
)

// Node represents one path stem in the project. It contains data derived from a YAML and/or SQL file (e.g. "/path/to/file.yaml" for "/path/to/file.sql").
type Node struct {
	Kind              ResourceKind
	Name              string
	Refs              []ResourceName
	Paths             []string
	YAML              *yaml.Node
	YAMLRaw           string
	YAMLPath          string
	Connector         string
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
		return p.parseMetricsView(ctx, node)
	case ResourceKindMigration:
		return p.parseMigration(ctx, node)
	default:
		panic(fmt.Errorf("unexpected resource kind: %s", node.Kind.String()))
	}
}

// commonYAML parses YAML fields common to all YAML files.
type commonYAML struct {
	// Kind can be inferred from the directory name in certain cases, but otherwise must be specified manually.
	Kind *string `yaml:"kind"`
	// Name is usually inferred from the filename, but can be specified manually.
	Name string `yaml:"name"`
	// Refs are a list of other resources that this resource depends on. They are usually inferred from other fields, but can also be specified manually.
	Refs []*yaml.Node `yaml:"refs"`
	// ParserConfig enables setting file-level parser config.
	ParserConfig struct {
		Templating *bool `yaml:"templating"`
	} `yaml:"parser"`
	// Connector names the connector to use for this resource. It may not be used in all resources, but is included here since it provides context for the SQL field.
	Connector string `yaml:"connector"`
	// SQL contains the SQL string for this resource. It may be specified inline, or will be loaded from a file at the same stem. It may not be supported in all resources.
	SQL string `yaml:"sql"`
}

// parseStem parses a pair of YAML and SQL files with the same path stem (e.g. "/path/to/file.yaml" for "/path/to/file.sql").
// Note that either of the YAML or SQL files may be empty (the paths arg will only contain non-nil paths).
func (p *Parser) parseStem(ctx context.Context, paths []string, ymlPath, yml, sqlPath, sql string) (*Node, error) {
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
		// Copy basic properties
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

		// Parse resource kind if set in YAML
		if cfg.Kind != nil {
			res.Kind, err = ParseResourceKind(*cfg.Kind)
			if err != nil {
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
		res.Refs = append(res.Refs, meta.Refs...) // If needed, deduplication happens in upsertResource

		// Additionally parse annotations provided in comments (e.g. "-- @materialize: true")
		commentAnnotations := sqlparse.ExtractAnnotations(res.SQL)
		if len(commentAnnotations) > 0 && res.SQLAnnotations == nil {
			res.SQLAnnotations = make(map[string]any)
		}
		for k, v := range commentAnnotations {
			res.SQLAnnotations[k] = v
		}

		// Expand dots in annotations. E.g. turn annotations["foo.bar"] into annotations["foo"]["bar"].
		res.SQLAnnotations, err = expandAnnotations(res.SQLAnnotations, "")
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
		case "kind":
			v, ok := v.(string)
			if !ok {
				err = fmt.Errorf("invalid type %T for property 'kind'", v)
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
			res.Name = fileutil.Stem(ymlPath)
		} else if sqlPath != "" {
			res.Name = fileutil.Stem(sqlPath)
		}
	}

	// If resource kind is not set in YAML or SQL, try to infer it from the context
	if res.Kind == ResourceKindUnspecified {
		if strings.HasPrefix(paths[0], "/sources") {
			res.Kind = ResourceKindSource
		} else if strings.HasPrefix(paths[0], "/models") {
			res.Kind = ResourceKindModel
		} else if strings.HasPrefix(paths[0], "/dashboards") {
			res.Kind = ResourceKindMetricsView
		} else if strings.HasPrefix(paths[0], "/init.sql") {
			res.Kind = ResourceKindMigration
		} else {
			path := ymlPath
			if path == "" {
				path = sqlPath
			}
			return nil, pathError{path: path, err: errors.New("resource kind not specified and could not be inferred from context")}
		}
	}

	return res, nil
}

// parseYAMLRefs parses a list of YAML nodes into a list of ResourceNames.
// It's used to parse the "refs" field in baseConfig.
func parseYAMLRefs(refs []*yaml.Node) ([]ResourceName, error) {
	var res []ResourceName
	for _, ref := range refs {
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

		// We support map refs of the form { kind: "kind", name: "my-resource" }
		if ref.Kind == yaml.MappingNode {
			var name ResourceName
			err := ref.Decode(&name)
			if err != nil {
				return nil, fmt.Errorf("invalid refs: %w", err)
			}
			res = append(res, name)
			continue
		}

		// ref was neither a string nor a map
		return nil, fmt.Errorf("invalid refs: %v", ref)
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
func expandAnnotations(annotations map[string]any, prefix string) (map[string]any, error) {
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
