package runtime

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/observability"
	"github.com/rilldata/rill/runtime/services/catalog"
	"go.uber.org/zap"
)

var errConnectionCacheClosed = errors.New("connectionCache: closed")

const _migrateTimeout = 30 * time.Second

// cache for instance specific connections only
// all instance specific connections should be opened via connection cache only
type connectionCache struct {
	lruCache *simplelru.LRU // items with zero references(opened but not in-use) ready for eviction
	cache    map[any]any    // items with non zero references (in-use) which should not be evicted
	lock     sync.Mutex
	closed   bool
	logger   *zap.Logger
}

type connWithRef struct {
	drivers.Handle
	ref int
}

func newConnectionCache(size int, logger *zap.Logger) *connectionCache {
	cache, err := simplelru.NewLRU(size, func(key interface{}, value interface{}) {
		// close the evicted connection
		if value.(*connWithRef).ref != 0 { // the callback also gets called when removing items manually i.e. transferring to in-use cache
			return
		}
		if err := value.(drivers.Handle).Close(); err != nil {
			logger.Error("failed closing cached connection for ", zap.String("key", key.(string)), zap.Error(err))
		}
	})
	if err != nil {
		panic(err)
	}
	return &connectionCache{lruCache: cache, cache: make(map[any]any), logger: logger}
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
	for _, key := range c.lruCache.Keys() {
		val, _ := c.lruCache.Get(key)
		err := val.(drivers.Handle).Close()
		if err != nil {
			c.logger.Error("failed closing cached connection", zap.Error(err))
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (c *connectionCache) get(ctx context.Context, instanceID, driver, dsn string, shared bool) (drivers.Handle, func(), error) {
	// TODO: This locks for all instances for the duration of Open and Migrate.
	// Adapt to lock only on the lookup, and then on the individual instance's Open and Migrate.

	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return nil, nil, errConnectionCacheClosed
	}

	key := instanceID + driver + dsn
	val, ok := c.cache[key]
	if !ok { // not in use
		val, ok = c.lruCache.Get(key)
		if !ok { // not opened
			logger := c.logger
			if instanceID != "default" {
				logger = c.logger.With(zap.String("instance_id", instanceID), zap.String("driver", driver))
			}
			conn, err := drivers.Open(driver, map[string]any{"dsn": dsn}, shared, logger)
			if err != nil {
				return nil, nil, err
			}

			ctx, cancel := context.WithTimeout(ctx, _migrateTimeout)
			defer cancel()

			err = conn.Migrate(ctx)
			if err != nil {
				conn.Close()
				return nil, nil, err
			}
			val = &connWithRef{Handle: conn, ref: 0}
		}
	}

	conn := val.(*connWithRef)
	// increase reference
	conn.ref += 1
	// transfer from lru to in-use cache
	c.cache[key] = val
	c.lruCache.Remove(key)
	return val.(drivers.Handle), func() {
		c.lock.Lock()
		defer c.lock.Unlock()

		conn.ref -= 1
		if conn.ref == 0 { // not in use
			// add key to lrucache for eviction
			c.lruCache.Add(key, val)
			// delete from in-use cache
			delete(c.cache, key)
		}
	}, nil
}

// evict removes the connection from cache and closes the connection
func (c *connectionCache) evict(ctx context.Context, instanceID, driver, dsn string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.closed {
		return false
	}

	key := instanceID + driver + dsn
	conn, ok := c.lruCache.Get(key)
	if !ok {
		conn, ok = c.cache[key]
	}
	if ok {
		err := conn.(drivers.Handle).Close()
		if err != nil {
			c.logger.Error("connection cache: failed to close cached connection", zap.Error(err), zap.String("instance", instanceID), observability.ZapCtx(ctx))
		}
		c.lruCache.Remove(key)
		delete(c.cache, key)
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
