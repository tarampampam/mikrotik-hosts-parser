package cache

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cacher is a byte-based cache with TTL.
type Cacher interface {
	// Get value associated with the key from the storage.
	Get(key string) (found bool, data []byte, ttl time.Duration, err error)

	// Put value into the storage.
	Put(key string, data []byte) error

	// Delete value from the storage with passed key.
	Delete(key string) (bool, error)
}

// Connector is a wrapper around the cache implementation.
type Connector interface {
	io.Closer

	// Open cache storage.
	Open() error

	// Cache returns cache implementation if Connector was successfully opened before.
	Cache() (Cacher, error)
}

// RedisConnector is Connector that uses RedisConnector under the hood.
type RedisConnector struct {
	ctx   context.Context
	opt   *redis.Options
	rdb   *redis.Client
	ttl   time.Duration
	cache *RedisCache
}

// NewRedisConnector creates new RedisConnector.
func NewRedisConnector(ctx context.Context, opt *redis.Options, ttl time.Duration) *RedisConnector {
	return &RedisConnector{ctx: ctx, opt: opt, ttl: ttl}
}

// Open establish connection to the redis server and verify it using Ping command.
func (c *RedisConnector) Open() error {
	if c.rdb != nil {
		return errors.New("already opened")
	}

	rdb := redis.NewClient(c.opt).WithContext(c.ctx)

	if err := rdb.Ping(c.ctx).Err(); err != nil {
		return err
	}

	c.rdb = rdb
	c.cache = NewRedisCache(c.ctx, c.rdb, c.ttl)

	return nil
}

// Test verifies connection to the redis server using Ping command.
func (c *RedisConnector) Test() error {
	if c.rdb == nil {
		return errors.New("not opened")
	}

	return c.rdb.Ping(c.ctx).Err()
}

// Close drops connection to the redis server and forgets about him.
func (c *RedisConnector) Close() error {
	if c.rdb == nil {
		return errors.New("already closed or was not opened")
	}

	defer func() { c.rdb, c.cache = nil, nil }()

	return c.rdb.Close()
}

// Cache returns cache implementation if RedisConnector was successfully opened before.
func (c *RedisConnector) Cache() (Cacher, error) {
	if c.rdb == nil || c.cache == nil {
		return nil, errors.New("not opened")
	}

	return c.cache, nil
}

// InMemoryConnector is Connector that uses InMemoryCache under the hood.
type InMemoryConnector struct {
	ttl, cleanupInterval time.Duration
	cache                *InMemoryCache
}

// NewInMemoryConnector creates new InMemoryConnector.
func NewInMemoryConnector(ttl, cleanupInterval time.Duration) *InMemoryConnector {
	return &InMemoryConnector{ttl: ttl, cleanupInterval: cleanupInterval}
}

// Open creates inmemory cache implementation.
func (c *InMemoryConnector) Open() error {
	if c.cache != nil {
		return errors.New("already opened")
	}

	c.cache = NewInMemoryCache(c.ttl, c.cleanupInterval)

	return nil
}

// Close closes current caching storage and forgets about him.
func (c *InMemoryConnector) Close() error {
	if c.cache == nil {
		return errors.New("already closed or was not opened")
	}

	defer func() { c.cache = nil }()

	return c.cache.Close()
}

// Cache returns cache implementation if InMemoryConnector was successfully opened before.
func (c *InMemoryConnector) Cache() (Cacher, error) {
	if c.cache == nil {
		return nil, errors.New("not opened")
	}

	return c.cache, nil
}
