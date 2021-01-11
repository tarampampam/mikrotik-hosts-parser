package cache

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
)

// Cacher is a cache engine.
type Cacher interface {
	// Get retrieves value for the key from the storage.
	Get(key string) (found bool, data []byte, ttl time.Duration, err error)

	// Put value into the storage.
	Put(key string, data []byte) error

	// Delete value from the storage with passed key.
	Delete(key string) (bool, error)
}

type Tester interface {
	Test() error
}

type Engine interface {
	io.Closer
	Open() error
	Cache() (Cacher, error)
}

type RedisEngine struct {
	ctx   context.Context
	opt   *redis.Options
	rdb   *redis.Client
	ttl   time.Duration
	cache *RedisCache
}

func NewRedisEngine(ctx context.Context, opt *redis.Options, ttl time.Duration) (*RedisEngine, error) {
	return &RedisEngine{ctx: ctx, opt: opt, ttl: ttl}, nil
}

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

func (e *RedisEngine) Test() error {
	if e.rdb == nil {
		return errors.New("not opened")
	}

	return e.rdb.Ping(e.ctx).Err()
}

func (e *RedisEngine) Close() error {
	if e.rdb == nil {
		return errors.New("already closed or was not opened")
	}

	defer func() { e.rdb, e.cache = nil, nil }()

	return e.rdb.Close()
}

func (e *RedisEngine) Cache() (Cacher, error) {
	if e.rdb == nil || e.cache == nil {
		return nil, errors.New("not opened")
	}

	return e.cache, nil
}

type InMemoryEngine struct {
	ttl, cleanupInterval time.Duration
	cache                *InMemoryCache
}

func NewInMemoryEngine(ttl, cleanupInterval time.Duration) (*InMemoryEngine, error) {
	return &InMemoryEngine{ttl: ttl, cleanupInterval: cleanupInterval}, nil
}

func (e *InMemoryEngine) Open() error {
	if e.cache != nil {
		return errors.New("already opened")
	}

	e.cache = NewInMemoryCache(e.ttl, e.cleanupInterval)

	return nil
}

func (e *InMemoryEngine) Close() error {
	if e.cache == nil {
		return errors.New("already closed or was not opened")
	}

	defer func() { e.cache = nil }()

	return e.cache.Close()
}

func (e *InMemoryEngine) Cache() (Cacher, error) {
	if e.cache == nil {
		return nil, errors.New("not opened")
	}

	return e.cache, nil
}
