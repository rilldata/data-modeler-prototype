package runtime

import (
	"context"
	"errors"
	"sync"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/services/catalog"
	"go.uber.org/zap"
)

var errConnectionCacheClosed = errors.New("connectionCache: closed")

// cache for instance specific connections only
// all instance specific connections should be opened via connection cache only
type connectionCache struct {
	cache  *simplelru.LRU
	lock   sync.Mutex
	closed bool
	logger *zap.Logger
}

func newConnectionCache(size int, logger *zap.Logger) *connectionCache {
	cache, err := simplelru.NewLRU(size, func(key interface{}, value interface{}) {
		// close the evicted connection
		if err := value.(drivers.Connection).Close(); err != nil {
			logger.Error("failed closing cached connection for ", zap.String("key", key.(string)), zap.Error(err))
		}
	})
	if err != nil {
		panic(err)
	}
	return &connectionCache{cache: cache, logger: logger}
}

func (c *connectionCache) Close() error {
	c.lock.Lock()
	if c.closed {
		c.lock.Unlock()
		return errConnectionCacheClosed
	}
	c.closed = true
	c.lock.Unlock()

	var firstErr error
	for _, key := range c.cache.Keys() {
		val, _ := c.cache.Get(key)
		err := val.(drivers.Connection).Close()
		if err != nil {
			c.logger.Error("failed closing cached connection", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (c *connectionCache) get(ctx context.Context, instanceID, driver, dsn string) (drivers.Connection, error) {
	// TODO: This locks for all instances for the duration of Open and Migrate.
	// Adapt to lock only on the lookup, and then on the individual instance's Open and Migrate.

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return nil, errConnectionCacheClosed
	}

	key := instanceID + driver + dsn
	val, ok := c.cache.Get(key)
	if !ok {
		logger := c.logger
		if instanceID != "default" {
			logger = c.logger.With(zap.String("instance_id", instanceID), zap.String("driver", driver))
		}
		conn, err := drivers.Open(driver, dsn, logger)
		if err != nil {
			return nil, err
		}

		err = conn.Migrate(ctx)
		if err != nil {
			return nil, err
		}

		c.cache.Add(key, conn)
		return conn, nil
	}

	return val.(drivers.Connection), nil
}

// evict removes the connection from cache and closes the connection
func (c *connectionCache) evict(ctx context.Context, instanceID, driver, dsn string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return false
	}

	key := instanceID + driver + dsn
	conn, ok := c.cache.Get(key)
	if ok {
		err := conn.(drivers.Connection).Close()
		if err != nil {
			c.logger.Error("connection cache: failed to close cached connection", zap.Error(err), zap.String("instance", instanceID), observability.ZapCtx(ctx))
		}
		c.cache.Remove(key)
	}
	return ok
}

type migrationMetaCache struct {
	cache *simplelru.LRU
	lock  sync.Mutex
}

func newMigrationMetaCache(size int) *migrationMetaCache {
	cache, err := simplelru.NewLRU(size, nil)
	if err != nil {
		panic(err)
	}

	return &migrationMetaCache{cache: cache}
}

func (c *migrationMetaCache) get(instID string) *catalog.MigrationMeta {
	c.lock.Lock()
	defer c.lock.Unlock()
	if val, ok := c.cache.Get(instID); ok {
		return val.(*catalog.MigrationMeta)
	}

	meta := catalog.NewMigrationMeta()
	c.cache.Add(instID, meta)
	return meta
}

func (c *migrationMetaCache) evict(ctx context.Context, instID string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cache.Remove(instID)
}
