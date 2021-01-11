package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache_GetPutDelete(t *testing.T) {
	mini, err := miniredis.Run()
	assert.NoError(t, err)

	defer mini.Close()

	cache := NewRedisCache(context.Background(), redis.NewClient(&redis.Options{Addr: mini.Addr()}), time.Minute)

	const testKeyName = "foo"

	// try to get non-existing entry
	found, data, ttl, err := cache.Get(testKeyName)
	assert.False(t, found)
	assert.Nil(t, data)
	assert.Zero(t, ttl)
	assert.NoError(t, err)

	// put valid value with the same key
	assert.NoError(t, cache.Put(testKeyName, []byte{1, 2, 3}))

	// and now all must be fine
	found, data, ttl, err = cache.Get(testKeyName)
	assert.True(t, found)
	assert.Equal(t, []byte{1, 2, 3}, data)
	assert.InDelta(t, time.Minute.Milliseconds(), ttl.Milliseconds(), 3)
	assert.NoError(t, err)

	// delete the key
	deleted, err := cache.Delete(testKeyName)
	assert.True(t, deleted)
	assert.NoError(t, err)

	// try to delete non-existing key
	deleted, err = cache.Delete(testKeyName)
	assert.False(t, deleted)
	assert.NoError(t, err)
}

func TestRedisCache_GetWithEmptyKey(t *testing.T) {
	cache := NewRedisCache(context.Background(), nil, time.Minute)

	found, data, ttl, err := cache.Get("")
	assert.False(t, found)
	assert.Nil(t, data)
	assert.Zero(t, ttl)
	assert.Error(t, err)
}

func TestRedisCache_PutWithEmptyKey(t *testing.T) {
	cache := NewRedisCache(context.Background(), nil, time.Minute)

	err := cache.Put("", []byte{1})
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyKey, err)
}

func TestRedisCache_PutWithEmptyData(t *testing.T) {
	cache := NewRedisCache(context.Background(), nil, time.Minute)

	err := cache.Put("foo", []byte{})
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyData, err)

	err = cache.Put("foo", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyData, err)
}

func TestRedisCache_DeleteWithEmptyKey(t *testing.T) {
	cache := NewRedisCache(context.Background(), nil, time.Minute)

	deleted, err := cache.Delete("")
	assert.False(t, deleted)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyKey, err)
}
