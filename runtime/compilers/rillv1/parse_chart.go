package rillv1

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ChartYaml struct {
	commonYAML `yaml:",inline"` // Not accessed here, only setting it so we can use KnownFields for YAML parsing
	Title      string           `yaml:"title"`
	Data       *DataYAML        `yaml:"data"`
	VegaLite   string           `yaml:"vega_lite"`
}

func (p *Parser) parseChart(node *Node) error {
	// Parse YAML
	tmp := &ChartYaml{}
	err := p.decodeNodeYAML(node, true, tmp)
	if err != nil {
		return err
	}

	// Validate SQL or connector isn't set
	if node.SQL != "" {
		return fmt.Errorf("charts cannot have SQL")
	}
	if !node.ConnectorInferred && node.Connector != "" {
		return fmt.Errorf("charts cannot have a connector")
	}

	if tmp.VegaLite == "" {
		return errors.New(`missing vega_lite configuration`)
	}
	if !json.Valid([]byte(tmp.VegaLite)) {
		return errors.New(`failed to parse "vega_lite" as JSON`)
	}

	if tmp.Data == nil {
		return errors.New(`missing required property "data"`)
	}
	resolver, resolverProps, resolverRefs, err := p.parseDataYAML(tmp.Data)
	if err != nil {
		return err
	}
	node.Refs = append(node.Refs, resolverRefs...)

	// Track chart
	r, err := p.insertResource(ResourceKindChart, node.Name, node.Paths, node.Refs...)
	if err != nil {
		return err
	}
	// NOTE: After calling insertResource, an error must not be returned. Any validation should be done before calling it.

	r.ChartSpec.Title = tmp.Title
	r.ChartSpec.Resolver = resolver
	r.ChartSpec.ResolverProperties = resolverProps
	r.ChartSpec.VegaLiteSpec = tmp.VegaLite

	return nil
}
