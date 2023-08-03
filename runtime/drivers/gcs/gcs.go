package gcs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/runtime/drivers"
	rillblob "github.com/rilldata/rill/runtime/drivers/blob"
	"github.com/rilldata/rill/runtime/pkg/fileutil"
	"github.com/rilldata/rill/runtime/pkg/gcputil"
	"github.com/rilldata/rill/runtime/pkg/globutil"
	"go.uber.org/zap"
	"gocloud.dev/blob/gcsblob"
	"gocloud.dev/gcp"
	"golang.org/x/oauth2"
	"google.golang.org/api/googleapi"
)

const defaultPageSize = 20

func init() {
	drivers.Register("gcs", driver{})
	drivers.RegisterAsConnector("gcs", driver{})
}

var spec = drivers.Spec{
	DisplayName:        "Google Cloud Storage",
	Description:        "Connect to Google Cloud Storage.",
	ServiceAccountDocs: "https://docs.rilldata.com/deploy/credentials/gcs",
	SourceProperties: []drivers.PropertySchema{
		{
			Key:         "path",
			DisplayName: "GS URI",
			Description: "Path to file on the disk.",
			Placeholder: "gs://bucket-name/path/to/file.csv",
			Type:        drivers.StringPropertyType,
			Required:    true,
			Hint:        "Glob patterns are supported",
		},
		{
			Key:         "gcp.credentials",
			DisplayName: "GCP credentials",
			Description: "GCP credentials inferred from your local environment.",
			Type:        drivers.InformationalPropertyType,
			Hint:        "Set your local credentials: <code>gcloud auth application-default login</code> Click to learn more.",
			Href:        "https://docs.rilldata.com/develop/import-data#configure-credentials-for-gcs",
		},
	},
	ConfigProperties: []drivers.PropertySchema{
		{
			Key:  "google_application_credentials",
			Hint: "Enter path of file to load from.",
			ValidateFunc: func(any interface{}) error {
				val := any.(string)
				if val == "" {
					// user can chhose to leave empty for public sources
					return nil
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}

				_, err = os.Stat(path)
				return err
			},
			TransformFunc: func(any interface{}) interface{} {
				val := any.(string)
				if val == "" {
					return ""
				}

				path, err := fileutil.ExpandHome(strings.TrimSpace(val))
				if err != nil {
					return err
				}
				// ignoring error since PathError is already validated
				content, _ := os.ReadFile(path)
				return string(content)
			},
		},
	},
}

type driver struct{}

type configProperties struct {
	SecretJSON      string `mapstructure:"google_application_credentials"`
	AllowHostAccess bool   `mapstructure:"allow_host_access"`
}

func (d driver) Open(config map[string]any, logger *zap.Logger) (drivers.Connection, error) {
	conf := &configProperties{}
	err := mapstructure.Decode(config, conf)
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		config: conf,
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
	b, ok := src.BucketSource()
	if !ok {
		return false, fmt.Errorf("require bucket source")
	}

	conf, err := parseSourceProperties(b.Properties)
	if err != nil {
		return false, fmt.Errorf("failed to parse config: %w", err)
	}

	client := gcp.NewAnonymousHTTPClient(gcp.DefaultTransport())
	bucketObj, err := gcsblob.OpenBucket(ctx, client, conf.url.Host, nil)
	if err != nil {
		return false, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	return bucketObj.IsAccessible(ctx)
}

type sourceProperties struct {
	Path                  string `key:"path"`
	GlobMaxTotalSize      int64  `mapstructure:"glob.max_total_size"`
	GlobMaxObjectsMatched int    `mapstructure:"glob.max_objects_matched"`
	GlobMaxObjectsListed  int64  `mapstructure:"glob.max_objects_listed"`
	GlobPageSize          int    `mapstructure:"glob.page_size"`
	url                   *globutil.URL
}

func parseSourceProperties(props map[string]any) (*sourceProperties, error) {
	conf := &sourceProperties{}
	err := mapstructure.Decode(props, conf)
	if err != nil {
		return nil, err
	}
	if !doublestar.ValidatePattern(conf.Path) {
		// ideally this should be validated at much earlier stage
		// keeping it here to have gcs specific validations
		return nil, fmt.Errorf("glob pattern %s is invalid", conf.Path)
	}
	url, err := globutil.ParseBucketURL(conf.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse path %q, %w", conf.Path, err)
	}

	if url.Scheme != "gs" {
		return nil, fmt.Errorf("invalid gcs path %q, should start with gs://", conf.Path)
	}

	conf.url = url
	return conf, nil
}

type Connection struct {
	config *configProperties
	logger *zap.Logger
}

var _ drivers.Connection = &Connection{}

// Driver implements drivers.Connection.
func (c *Connection) Driver() string {
	return "gcs"
}

// Config implements drivers.Connection.
func (c *Connection) Config() map[string]any {
	m := make(map[string]any, 0)
	_ = mapstructure.Decode(c.config, m)
	return m
}

// Close implements drivers.Connection.
func (c *Connection) Close() error {
	return nil
}

// Registry implements drivers.Connection.
func (c *Connection) AsRegistry() (drivers.RegistryStore, bool) {
	return nil, false
}

// Catalog implements drivers.Connection.
func (c *Connection) AsCatalogStore() (drivers.CatalogStore, bool) {
	return nil, false
}

// Repo implements drivers.Connection.
func (c *Connection) AsRepoStore() (drivers.RepoStore, bool) {
	return nil, false
}

// OLAP implements drivers.Connection.
func (c *Connection) AsOLAP() (drivers.OLAPStore, bool) {
	return nil, false
}

// Migrate implements drivers.Connection.
func (c *Connection) Migrate(ctx context.Context) (err error) {
	return nil
}

// MigrationStatus implements drivers.Connection.
func (c *Connection) MigrationStatus(ctx context.Context) (current, desired int, err error) {
	return 0, 0, nil
}

// AsObjectStore implements drivers.Connection.
func (c *Connection) AsObjectStore() (drivers.ObjectStore, bool) {
	return c, true
}

// AsTransporter implements drivers.Connection.
func (c *Connection) AsTransporter(from, to drivers.Connection) (drivers.Transporter, bool) {
	return nil, false
}

func (c *Connection) AsFileStore() (drivers.FileStore, bool) {
	return nil, false
}

// AsSQLStore implements drivers.Connection.
func (c *Connection) AsSQLStore() (drivers.SQLStore, bool) {
	return nil, false
}

// DownloadFiles returns a file iterator over objects stored in gcs.
// The credential json is read from config google_application_credentials.
// Additionally in case `allow_host_credentials` is true it looks for "Application Default Credentials" as well
func (c *Connection) DownloadFiles(ctx context.Context, source *drivers.BucketSource) (drivers.FileIterator, error) {
	conf, err := parseSourceProperties(source.Properties)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	client, err := c.createClient(ctx)
	if err != nil {
		return nil, err
	}

	bucketObj, err := gcsblob.OpenBucket(ctx, client, conf.url.Host, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket %q, %w", conf.url.Host, err)
	}

	// prepare fetch configs
	opts := rillblob.Options{
		GlobMaxTotalSize:      conf.GlobMaxTotalSize,
		GlobMaxObjectsMatched: conf.GlobMaxObjectsMatched,
		GlobMaxObjectsListed:  conf.GlobMaxObjectsListed,
		GlobPageSize:          conf.GlobPageSize,
		GlobPattern:           conf.url.Path,
		ExtractPolicy:         source.ExtractPolicy,
	}

	iter, err := rillblob.NewIterator(ctx, bucketObj, opts, c.logger)
	if err != nil {
		apiError := &googleapi.Error{}
		// in cases when no creds are passed
		if errors.As(err, &apiError) && apiError.Code == http.StatusUnauthorized {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", apiError))
		}

		// StatusUnauthorized when incorrect key is passsed
		// StatusBadRequest when key doesn't have a valid credentials file
		retrieveError := &oauth2.RetrieveError{}
		if errors.As(err, &retrieveError) && (retrieveError.Response.StatusCode == http.StatusUnauthorized || retrieveError.Response.StatusCode == http.StatusBadRequest) {
			return nil, drivers.NewPermissionDeniedError(fmt.Sprintf("can't access remote err: %v", retrieveError))
		}
	}

	return iter, err
}

func (c *Connection) createClient(ctx context.Context) (*gcp.HTTPClient, error) {
	creds, err := gcputil.Credentials(ctx, c.config.SecretJSON, c.config.AllowHostAccess)
	if err != nil {
		if !errors.Is(err, gcputil.ErrNoCredentials) {
			return nil, err
		}

		// no credentials set, we try with a anonymous client in case user is trying to access public buckets
		return gcp.NewAnonymousHTTPClient(gcp.DefaultTransport()), nil
	}
	// the token source returned from credentials works for all kind of credentials like serviceAccountKey, credentialsKey etc.
	return gcp.NewHTTPClient(gcp.DefaultTransport(), gcp.CredentialsTokenSource(creds))
}
