package cache

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisCache is a redis cache implementation.
type RedisCache struct {
	ctx   context.Context
	redis *redis.Client
	ttl   time.Duration
}

// NewRedisCache creates new redis cache instance.
func NewRedisCache(ctx context.Context, client *redis.Client, ttl time.Duration) *RedisCache {
	return &RedisCache{ctx: ctx, redis: client, ttl: ttl}
}

// key generates cache entry key using passed string.
func (c *RedisCache) key(s string) string {
	h := md5.Sum([]byte(s)) //nolint:gosec

	return "cache:" + hex.EncodeToString(h[:])
}

// TTL returns current cache values time-to-live.
func (c *RedisCache) TTL() time.Duration { return c.ttl }

// Get retrieves value for the key from the storage.
func (c *RedisCache) Get(key string) (found bool, data []byte, ttl time.Duration, err error) {
	if key == "" {
		err = ErrEmptyKey

		return // wrong argument
	}

	k := c.key(key)

	if data, err = c.redis.Get(c.ctx, k).Bytes(); err != nil {
		if errors.Is(err, redis.Nil) {
			err = nil

			return // not found
		}

		return // key getting failed
	}

	found = true

	if ttl, err = c.redis.TTL(c.ctx, k).Result(); err != nil {
		return // ttl getting failed
	}

	return // all is ok
}

// Put value into the storage.
func (c *RedisCache) Put(key string, data []byte) error {
	if key == "" {
		return ErrEmptyKey
	} else if len(data) == 0 {
		return ErrEmptyData
	}

	return c.redis.Set(c.ctx, c.key(key), data, c.ttl).Err()
}

// Delete value from the storage with passed key.
func (c *RedisCache) Delete(key string) (bool, error) {
	if key == "" {
		return false, ErrEmptyKey
	}

	if count, err := c.redis.Del(c.ctx, c.key(key)).Result(); err != nil {
		return false, err
	} else if count <= 0 {
		return false, nil
	}

	return true, nil
}
