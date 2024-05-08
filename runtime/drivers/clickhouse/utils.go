package clickhouse

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
)

type sinkProperties struct {
	Table string `mapstructure:"table"`
}

func parseSinkProperties(props map[string]any) (*sinkProperties, error) {
	cfg := &sinkProperties{}
	if err := mapstructure.Decode(props, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse sink properties: %w", err)
	}
	return cfg, nil
}

func safeSQLName(name string) string {
	if name == "" {
		return name
	}
	return fmt.Sprintf("\"%s\"", strings.ReplaceAll(name, "\"", "\"\""))
}
