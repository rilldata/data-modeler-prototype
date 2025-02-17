package rillv1

import (
	"context"
	"slices"
	"strings"

	runtimev1 "github.com/rilldata/rill/proto/gen/rill/runtime/v1"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/drivers/slack"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"google.golang.org/protobuf/types/known/structpb"
)

// Connector contains metadata about a connector used in a Rill project
type Connector struct {
	Name            string
	Driver          string
	Spec            *drivers.Spec
	DefaultConfig   map[string]string
	Resources       []*Resource
	AnonymousAccess bool
	Err             error
}

// AnalyzeConnectors extracts connector metadata from a Rill project
func (p *Parser) AnalyzeConnectors(ctx context.Context) []*Connector {
	a := &connectorAnalyzer{
		parser: p,
		result: make(map[string]*Connector),
	}

	a.analyze(ctx)

	res := maps.Values(a.result)

	// Sort output to ensure deterministic ordering
	slices.SortFunc(res, func(a, b *Connector) int {
		return strings.Compare(a.Name, b.Name)
	})

	return res
}

// connectorAnalyzer implements logic for extracting connector metadata from a parser
type connectorAnalyzer struct {
	parser *Parser
	result map[string]*Connector
}

// analyze is the entrypoint for connector analysis. After running it, you can access the result.
func (a *connectorAnalyzer) analyze(ctx context.Context) {
	if a.parser.RillYAML != nil {
		// Track any connectors explicitly configured in rill.yaml
		for _, c := range a.parser.RillYAML.Connectors {
			a.trackConnector(c.Name, nil, false)
		}

		// Track the OLAP connector specified in rill.yaml
		if a.parser.RillYAML.OLAPConnector != "" {
			a.trackConnector(a.parser.RillYAML.OLAPConnector, nil, false)
		}
	}

	for _, r := range a.parser.Resources {
		a.analyzeResource(ctx, r)
	}
}

// analyzeResource extracts connector metadata for a single resource.
// NOTE: If we add more resource kinds that use connectors, add connector extraction logic here.
func (a *connectorAnalyzer) analyzeResource(ctx context.Context, r *Resource) {
	if r.ModelSpec != nil {
		a.analyzeModel(ctx, r)
	} else if r.MetricsViewSpec != nil {
		a.trackConnector(r.MetricsViewSpec.Connector, r, false)
	} else if r.MigrationSpec != nil {
		a.trackConnector(r.MigrationSpec.Connector, r, false)
	} else if r.APISpec != nil {
		a.analyzeResourceWithResolver(r, r.APISpec.Resolver, r.APISpec.ResolverProperties)
	} else if r.AlertSpec != nil {
		a.analyzeResourceNotifiers(r, r.AlertSpec.Notifiers)
	} else if r.ReportSpec != nil {
		a.analyzeResourceNotifiers(r, r.ReportSpec.Notifiers)
	} else if r.ConnectorSpec != nil {
		// resource is not passed to prevent the connector depends on itself
		a.trackConnector(r.Name.Name, nil, false)
	}
	// Other resource kinds currently don't use connectors.
}

// analyzeModel extracts connector metadata for a model resource.
// The logic for extracting metadata from a model is more complex than for other resource kinds, hence the separate function.
func (a *connectorAnalyzer) analyzeModel(ctx context.Context, r *Resource) {
	// No analysis necessary for the output connector
	a.trackConnector(r.ModelSpec.OutputConnector, r, false)

	// Prep for analyzing InputConnector
	spec := r.ModelSpec
	inputProps := spec.InputProperties.AsMap()
	_, inputDriver, driverErr := a.parser.driverForConnector(spec.InputConnector)
	if driverErr != nil {
		// Track the errored input connector and return
		a.trackConnector(spec.InputConnector, r, false)
		return
	}

	// Check if we have anonymous access (unless we already know that we don't)
	var anonAccess bool
	if res, ok := a.result[spec.InputConnector]; !ok || res.AnonymousAccess {
		anonAccess, _ = inputDriver.HasAnonymousSourceAccess(ctx, inputProps, zap.NewNop())
	}

	// Track the input connector
	a.trackConnector(spec.InputConnector, r, anonAccess)

	if spec.StageConnector != "" {
		// Track the staging connector
		// We need write access to the stage connector so tracking without analysis
		a.trackConnector(spec.StageConnector, r, false)
	}

	// Track any tertiary connectors (like a DuckDB source referencing S3 in its SQL).
	// NOTE: Not checking anonymous access for these since we don't know what properties to use.
	// TODO: Can we solve that issue?
	otherConnectors, _ := inputDriver.TertiarySourceConnectors(ctx, inputProps, zap.NewNop())
	for _, connector := range otherConnectors {
		a.trackConnector(connector, r, false)
	}

	// Track the incremental state connector
	if spec.IncrementalStateResolver != "" && spec.IncrementalStateResolverProperties != nil {
		a.analyzeResourceWithResolver(r, spec.IncrementalStateResolver, spec.IncrementalStateResolverProperties)
	}
}

// analyzeResourceWithResolver extracts connector metadata for a resource that uses a resolver.
func (a *connectorAnalyzer) analyzeResourceWithResolver(r *Resource, resolver string, resolverProps *structpb.Struct) {
	// The "sql" and "glob" resolvers take an optional "connector" property
	if resolver == "sql" || resolver == "glob" {
		for k, v := range resolverProps.Fields {
			if k == "connector" {
				connector := v.GetStringValue()
				if connector != "" {
					a.trackConnector(connector, r, false)
					return
				}
			}
		}
	}
}

// analyzeResourceNotifiers extracts connector metadata for a resource that uses notifiers (email, slack, etc).
func (a *connectorAnalyzer) analyzeResourceNotifiers(r *Resource, notifiers []*runtimev1.Notifier) {
	for _, n := range notifiers {
		if n.Connector == "email" {
			// NOTE: email is not implemented as a real driver yet, so we skip it.
			continue
		}

		// Slack notifier can be used anonymously if no users and no channels are specified (only webhooks)
		anonAccess := false
		if n.Connector == "slack" {
			props, err := slack.DecodeProps(n.Properties.AsMap())
			if err == nil {
				if len(props.Users) == 0 && len(props.Channels) == 0 {
					anonAccess = true
				}
			}
		}

		a.trackConnector(n.Connector, r, anonAccess)
	}
}

// trackConnector tracks a connector and an associated resource in the analyzer's result map
func (a *connectorAnalyzer) trackConnector(connector string, r *Resource, anonAccess bool) {
	res, ok := a.result[connector]
	if !ok {
		// Search rill.yaml for default config properties for this connector
		var defaultConfig map[string]string
		if a.parser.RillYAML != nil {
			for _, c := range a.parser.RillYAML.Connectors {
				if c.Name == connector {
					defaultConfig = c.Defaults
					break
				}
			}
		}

		// Search among dedicated connectors
		for _, c := range a.parser.Resources {
			if c.ConnectorSpec != nil && c.Name.Name == connector {
				defaultConfig = c.ConnectorSpec.Properties
				break
			}
		}

		driver, driverConnector, driverErr := a.parser.driverForConnector(connector)
		if driverErr != nil {
			res = &Connector{
				Name: connector,
				Err:  driverErr,
			}
		} else {
			driverSpec := driverConnector.Spec()
			res = &Connector{
				Name:            connector,
				Driver:          driver,
				Spec:            &driverSpec,
				DefaultConfig:   defaultConfig,
				AnonymousAccess: true,
			}
		}

		a.result[connector] = res
	}

	if r != nil {
		found := false
		for _, existing := range res.Resources {
			if r.Name.Normalized() == existing.Name.Normalized() {
				found = true
				break
			}
		}
		if !found {
			res.Resources = append(res.Resources, r)
		}
	}

	if !anonAccess {
		res.AnonymousAccess = false
	}
}
