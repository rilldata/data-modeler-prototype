package https

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/connectors"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
)

func init() {
	connectors.Register("https", connector{})
}

var spec = connectors.Spec{
	DisplayName: "http(s)",
	Description: "Connect to a remote file.",
	Properties: []connectors.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Path",
			Description: "Path to the remote file.",
			Placeholder: "https://example.com/file.csv",
			Type:        connectors.StringPropertyType,
			Required:    true,
		},
	},
}

type Config struct {
	Path string `mapstructure:"path"`
}

func ParseConfig(props map[string]any) (*Config, error) {
	conf := &Config{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connector struct{}

func (c connector) Spec() connectors.Spec {
	return spec
}

func (c connector) ConsumeAsFile(ctx context.Context, source *connectors.Source) (string, error) {
	conf, err := ParseConfig(source.Properties)
	if err != nil {
		return "", fmt.Errorf("failed to parse config: %v", err)
	}

	extension, err := getUrlExtension(conf.Path)
	if err != nil {
		return "", fmt.Errorf("failed to parse path %s, %v", conf.Path, err)
	}

	resp, err := http.Get(conf.Path)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url %s:  %v", conf.Path, err)
	}
	defer resp.Body.Close()
	return fileutil.CopyToTempFile(resp.Body, source.Name, extension)
}

func getUrlExtension(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return fileutil.FullExt(u.Path), nil
}
