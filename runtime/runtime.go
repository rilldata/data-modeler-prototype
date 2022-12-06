package runtime

import (
	"context"
	"fmt"

	"github.com/rilldata/rill/runtime/drivers"
	"go.uber.org/zap"
)

type Options struct {
	ConnectionCacheSize int
	MetastoreDriver     string
	MetastoreDSN        string
}

type Runtime struct {
	opts         *Options
	metastore    drivers.Connection
	logger       *zap.Logger
	connCache    *connectionCache
	catalogCache *catalogCache
}

func New(opts *Options, logger *zap.Logger) (*Runtime, error) {
	// Open metadata db connection
	metastore, err := drivers.Open(opts.MetastoreDriver, opts.MetastoreDSN)
	if err != nil {
		return nil, fmt.Errorf("could not connect to metadata db: %w", err)
	}
	err = metastore.Migrate(context.Background())
	if err != nil {
		return nil, fmt.Errorf("metadata db migration: %w", err)
	}

	// Check the metastore is a registry
	_, ok := metastore.RegistryStore()
	if !ok {
		return nil, fmt.Errorf("server metastore must be a valid registry")
	}

	return &Runtime{
		opts:         opts,
		metastore:    metastore,
		logger:       logger,
		connCache:    newConnectionCache(opts.ConnectionCacheSize),
		catalogCache: newCatalogCache(),
	}, nil
}

func (rt *Runtime) Execute(ctx context.Context, instanceID string, priority int, sql string) (*drivers.Result, error) {
	// Get OLAP connection
	olap, err := rt.OLAP(ctx, instanceID)
	if err != nil {
		return nil, err
	}

	if olap.Dialect() != drivers.DialectDuckDB {
		return nil, fmt.Errorf("not available for dialect '%s'", olap.Dialect())
	}

	result, err := olap.Execute(ctx, &drivers.Statement{
		Query:    sql,
		Priority: priority,
	})
	return result, err
}
