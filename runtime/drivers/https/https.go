package https

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("https", driver{})
	drivers.RegisterAsConnector("https", driver{})
}

var spec = drivers.Spec{
	DisplayName: "http(s)",
	Description: "Connect to a remote file.",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "path",
			DisplayName: "Path",
			Description: "Path to the remote file.",
			Placeholder: "https://example.com/file.csv",
			Type:        drivers.StringPropertyType,
			Required:    true,
		},
	},
}

type driver struct{}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	conn := &connection{
		config: config,
		logger: logger,
	}
	return conn, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src drivers.Source, logger *zap.Logger) (bool, error) {
	return true, nil
}

type sourceProperties struct {
	Path    string            `mapstructure:"path"`
	Headers map[string]string `mapstructure:"headers"`
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

type connection struct {
	config map[string]any
	logger *zap.Logger
}

var _ drivers.Connection = &connection{}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "https"
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.config
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) AsRepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP() (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return nil, false
}

// AsTransporter implements drivers.Connection.
func (c *connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

// AsSQLStore implements drivers.Connection.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// FilePaths implements drivers.FileStore
func (c *connection) FilePaths(ctx context.Context, src *drivers.FileSource) ([]string, error) {
	conf, err := parseSourceProperties(src.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	extension, err := urlExtension(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %s, %w", conf.Path, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, conf.Path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url %s:  %w", conf.Path, err)
	}

	for k, v := range conf.Headers {
		req.Header.Set(k, v)
	}

	start := time.Now()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch url %s:  %w", conf.Path, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch url %s: %s", conf.Path, resp.Status)
	}

	// TODO :: I don't like src.Name
	file, size, err := fileutil.CopyToTempFile(resp.Body, src.Name, extension)
	if err != nil {
		return nil, err
	}

	// Collect metrics of download size and time
	drivers.RecordDownloadMetrics(ctx, &drivers.DownloadMetrics{
		Connector: "https",
		Ext:       extension,
		Duration:  time.Since(start),
		Size:      size,
	})

	return []string{file}, nil
}

func urlExtension(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	return fileutil.FullExt(u.Path), nil
}
