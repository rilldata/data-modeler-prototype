package druid

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"go.uber.org/zap"

	// Load Druid database/sql driver
	_ "github.com/rilldata/rill/runtime/drivers/druid/druidsqldriver"
)

var spec = drivers.Spec{
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:         "dsn",
			Type:        drivers.StringPropertyType,
			Required:    true,
			Description: "Druid connection string (using the Avatica protobuf endpoint)",
			Secret:      true,
		},
	},
}

func init() {
	drivers.Register("druid", &driver{})
	drivers.RegisterAsConnector("druid", &driver{})
}

type driver struct{}

var _ drivers.Driver = &driver{}

// Opens a connection to Apache Druid using HTTP API.
// Note that the Druid connection string must have the form "http://user:password@host:port/druid/v2/sql".
func (d *driver) Open(config map[string]any, shared bool, client *activity.Client, logger *zap.Logger) (drivers.Handle, error) {
	if shared {
		return nil, fmt.Errorf("druid driver can't be shared")
	}
	dsn, ok := config["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("require dsn to open druid connection")
	}

	dsn, err := correctURL(dsn)
	if err != nil {
		return nil, err
	}

	db, err := sqlx.Open("druid", dsn)
	if err != nil {
		return nil, err
	}

	// very roughly approximating num queries required for a typical page load
	db.SetMaxOpenConns(20)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("druid: %w", err)
	}

	conn := &connection{
		db:     db,
		config: config,
	}
	return conn, nil
}

func (d *driver) Drop(config map[string]any, logger *zap.Logger) error {
	return drivers.ErrDropNotSupported
}

func (d *driver) Spec() drivers.Spec {
	return spec
}

func (d *driver) HasAnonymousSourceAccess(ctx context.Context, src map[string]any, logger *zap.Logger) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (d *driver) TertiarySourceConnectors(ctx context.Context, src map[string]any, logger *zap.Logger) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}

type connection struct {
	db     *sqlx.DB
	config map[string]any
}

// Driver implements drivers.Connection.
func (c *connection) Driver() string {
	return "druid"
}

// Config used to open the Connection
func (c *connection) Config() map[string]any {
	return c.config
}

// Close implements drivers.Connection.
func (c *connection) Close() error {
	return c.db.Close()
}

// Registry implements drivers.Connection.
func (c *connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *connection) AsCatalogStore(instanceID string) (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *connection) AsRepoStore(instanceID string) (drivers.RepoStore, bool) {
	return nil, false
}

// AsAdmin implements drivers.Handle.
func (c *connection) AsAdmin(instanceID string) (drivers.AdminService, bool) {
	return nil, false
}

// AsAI implements drivers.Handle.
func (c *connection) AsAI(instanceID string) (drivers.AIService, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *connection) AsOLAP(instanceID string) (drivers.OLAPStore, bool) {
	return c, true
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
	return nil, false
}

// AsSQLStore implements drivers.Connection.
// Use OLAPStore instead.
func (c *connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

func (c *connection) EstimateSize() (int64, bool) {
	return 0, false
}

func (c *connection) AcquireLongRunning(ctx context.Context) (func(), error) {
	return func() {}, nil
}

func correctURL(dsn string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}

	if strings.Contains(u.Path, "avatica-protobuf") {
		avaticaUser := url.QueryEscape(u.Query().Get("avaticaUser"))
		avaticaPassword := url.QueryEscape(u.Query().Get("avaticaPassword"))

		if avaticaUser != "" {
			dsn = u.Scheme + "://" + avaticaUser + ":" + avaticaPassword + "@" + u.Host + "/druid/v2/sql"
		} else {
			dsn = u.Scheme + "://" + u.Host + "/druid/v2/sql"

		}
	}
	return dsn, nil
}
