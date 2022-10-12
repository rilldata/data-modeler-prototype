package connectors

import (
	"fmt"
)

// Connectors tracks all registered connector drivers
var Connectors = make(map[string]Connector)

// Register tracks a connector driver
func Register(name string, connector Connector) {
	if Connectors[name] != nil {
		panic(fmt.Errorf("already registered connector with name '%s'", name))
	}
	Connectors[name] = connector
}

// Connector is a driver for ingesting data from an external system
type Connector interface {
	Spec() Spec

	// TODO: Add method that extracts a source and outputs a schema and buffered
	// iterator for data in it. For consumption by a drivers.OLAPStore. Also consider
	// how to communicate splits and long-running/streaming data (e.g. for Kafka).
	// Consume(ctx context.Context, source Source) error
}

// Spec provides metadata about a connector and the properties it supports.
type Spec struct {
	DisplayName string
	Description string
	Properties  []PropertySchema
}

// PropertySchema provides the schema for a property supported by a connector.
type PropertySchema struct {
	Key         string
	Type        PropertySchemaType
	Required    bool
	DisplayName string
	Description string
	Placeholder string
}

// PropertySchemaType is an enum of types supported for connector properties.
type PropertySchemaType int

const (
	UnspecifiedPropertyType PropertySchemaType = iota
	StringPropertyType
	NumberPropertyType
	BooleanPropertyType
)

// Validate checks that val has the correct type
func (ps PropertySchema) ValidateType(val any) bool {
	switch val.(type) {
	case string:
		return ps.Type == StringPropertyType
	case bool:
		return ps.Type == BooleanPropertyType
	case int, int8, int16, int32, int64, float32, float64:
		return ps.Type == NumberPropertyType
	default:
		return false
	}
}

// Source represents a dataset to ingest using a specific connector (like a connector instance)
type Source struct {
	Name         string
	Connector    string
	SamplePolicy *SamplePolicy
	Properties   map[string]any
}

// SamplePolicy tells the connector to only ingest a sample of data from the source.
// Support for it is currently not implemented.
type SamplePolicy struct {
	Strategy string
	Sample   float32
	Limit    int
}

// Validate checks the source's properties against its connector's spec
func (s *Source) Validate() error {
	connector, ok := Connectors[s.Connector]
	if !ok {
		return fmt.Errorf("connector: not found")
	}

	for _, propSchema := range connector.Spec().Properties {
		val, ok := s.Properties[propSchema.Key]
		if !ok {
			if propSchema.Required {
				return fmt.Errorf("missing required property '%s'", propSchema.Key)
			}
			continue
		}

		if !propSchema.ValidateType(val) {
			return fmt.Errorf("unexpected type '%T' for property '%s'", val, propSchema.Key)
		}
	}

	return nil
}
