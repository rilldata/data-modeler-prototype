package file

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"go.uber.org/zap"
)

func init() {
	drivers.Register("file", driver{name: "file"})
	drivers.Register("local_file", driver{name: "local_file"})
	drivers.RegisterAsConnector("local_file", driver{name: "local_file"})
}

var spec = drivers.Spec{
	DisplayName: "Local file",
	Description: "Import Locally Stored File.",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "path",
			Type:        drivers.StringPropertyType,
			Required:    true,
			DisplayName: "Path",
			Description: "Path or URL to file",
			Placeholder: "/path/to/file",
		},
		{
			Key:         "format",
			Type:        drivers.StringPropertyType,
			Required:    false,
			DisplayName: "Format",
			Description: "Either CSV or Parquet. Inferred if not set.",
			Placeholder: "csv",
		},
	},
}

type driver struct {
	name string
}

func (d driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("file driver can't be shared")
	}
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open file connection")
	}

	path, err := fileutil.ExpandHome(dsn)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	c := &connection{
		root:         absPath,
		driverConfig: config,
		driverName:   d.name,
		shared:       shared,
	}
	if err := c.checkRoot(); err != nil {
		return nil, err
	}
	return c, nil
}

func (d driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d driver) Spec() drivers.Spec {
	return spec
}

func (d driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return true, nil
}

func (d driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, nil
}

type sourceProperties struct {
	Path   string `mapstructure:"path"`
	Format string `mapstructure:"format"`
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
	// root should be absolute path
	root         string
	driverConfig map[string]any
	driverName   string
	shared       bool

	watcherMu    sync.Mutex
	watcherCount int
	watcher      *watcher
}

// Config implements drivers.Connection.
func (c *connection) Config() map[string]any {
	return c.driverConfig
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return nil
}

// AsRegistry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// AsCatalogStore implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// AsRepoStore implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	if c.shared {
		return nil, false
	}
	return c, true
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// AsOLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
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
func (c *connection) AsTransporter(from, to drivers.Handle) (drivers.Transporter, bool) {
	return nil, false
}

// AsFileStore implements drivers.Connection.
func (c *connection) AsFileStore() (drivers.FileStore, bool) {
	return c, true
}

// AsSQLStore implements drivers.Connection.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// checkPath checks that the connection's root is a valid directory.
func (c *connection) checkRoot() error {
	info, err := os.Stat(c.root)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("repo: directory does not exist at '%s'", c.root)
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("repo: file is not a directory '%s'", c.root)
	}

	return nil
}
