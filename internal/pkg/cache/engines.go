package cache

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cacher is a byte-cache.
type Cacher interface {
	// Get value associated with the key from the storage.
	Get(key string) (found bool, data []byte, ttl time.Duration, err error)

	// Put value into the storage.
	Put(key string, data []byte) error

	// Delete value from the storage with passed key.
	Delete(key string) (bool, error)
}

// Tester allows to test something.
type Tester interface {
	// Test returns an error only when something is broken down inside.
	Test() error
}

// Engine is a wrapper around the cache implementation.
type Engine interface {
	io.Closer

	// Open cache storage.
	Open() error

	// Cache returns cache implementation if Engine was successfully opened before.
	Cache() (Cacher, error)
}

// RedisEngine is Engine that uses RedisEngine under the hood.
type RedisEngine struct {
	ctx   context.Context
	opt   *redis.Options
	rdb   *redis.Client
	ttl   time.Duration
	cache *RedisCache
}

// NewRedisEngine creates new RedisEngine.
func NewRedisEngine(ctx context.Context, opt *redis.Options, ttl time.Duration) *RedisEngine {
	return &RedisEngine{ctx: ctx, opt: opt, ttl: ttl}
}

// Open establish connection to the redis server and verify it using Ping command.
func (e *RedisEngine) Open() error {
	if e.rdb != nil {
		return errors.New("already opened")
	}

	rdb := redis.NewClient(e.opt).WithContext(e.ctx)

	if err := rdb.Ping(e.ctx).Err(); err != nil {
		return err
	}

	e.rdb = rdb
	e.cache = NewRedisCache(e.ctx, e.rdb, e.ttl)

	return nil
}

// Test verifies connection to the redis server using Ping command.
func (e *RedisEngine) Test() error {
	if e.rdb == nil {
		return errors.New("not opened")
	}

	return e.rdb.Ping(e.ctx).Err()
}

// Close drops connection to the redis server and forget about it.
func (e *RedisEngine) Close() error {
	if e.rdb == nil {
		return errors.New("already closed or was not opened")
	}

	defer func() { e.rdb, e.cache = nil, nil }()

	return e.rdb.Close()
}

// Cache returns cache implementation if RedisEngine was successfully opened before.
func (e *RedisEngine) Cache() (Cacher, error) {
	if e.rdb == nil || e.cache == nil {
		return nil, errors.New("not opened")
	}

	return e.cache, nil
}

// InMemoryEngine is Engine that uses InMemoryCache under the hood.
type InMemoryEngine struct {
	ttl, cleanupInterval time.Duration
	cache                *InMemoryCache
}

// NewInMemoryEngine creates new InMemoryEngine.
func NewInMemoryEngine(ttl, cleanupInterval time.Duration) *InMemoryEngine {
	return &InMemoryEngine{ttl: ttl, cleanupInterval: cleanupInterval}
}

// Open creates inmemory cache implementation.
func (e *InMemoryEngine) Open() error {
	if e.cache != nil {
		return errors.New("already opened")
	}

	e.cache = NewInMemoryCache(e.ttl, e.cleanupInterval)

	return nil
}

// Close closes current caching storage and forget about it.
func (e *InMemoryEngine) Close() error {
	if e.cache == nil {
		return errors.New("already closed or was not opened")
	}

	defer func() { e.cache = nil }()

	return e.cache.Close()
}

// Cache returns cache implementation if InMemoryEngine was successfully opened before.
func (e *InMemoryEngine) Cache() (Cacher, error) {
	if e.cache == nil {
		return nil, errors.New("not opened")
	}

	return e.cache, nil
}
