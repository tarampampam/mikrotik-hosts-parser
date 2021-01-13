package cache

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*
func BenchmarkInMemoryCache_Put(b *testing.B) {
	b.ReportAllocs()

	cache := NewInMemoryCache(time.Minute, time.Second)
	defer cache.Close()

	data := []byte(strings.Repeat("xxxxxxxxx", 102400))

	for n := 0; n < b.N; n++ {
		cache.Put("foo" + strconv.Itoa(n), data)
	}
}
*/

func TestInMemoryCache_GetPutDelete(t *testing.T) {
	cache := NewInMemoryCache(time.Minute, time.Second)
	defer cache.Close()

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

func TestInMemoryCache_CloserInterface(t *testing.T) {
	var cache Cacher //nolint:gosimple

	cache = NewInMemoryCache(time.Minute, time.Second)
	defer cache.(io.Closer).Close()

	_, ok := cache.(io.Closer)

	assert.True(t, ok)
}

func TestInMemoryCache_Expiration(t *testing.T) {
	const testKeyName = "foo"

	cache := NewInMemoryCache(time.Millisecond*100, time.Millisecond)
	defer cache.Close()

	assert.NoError(t, cache.Put(testKeyName, []byte{1, 2, 3}))

	found, _, _, _ := cache.Get(testKeyName) //nolint:dogsled
	assert.True(t, found)

	<-time.After(time.Millisecond * 98)

	found, _, _, _ = cache.Get(testKeyName) //nolint:dogsled
	assert.True(t, found)

	<-time.After(time.Millisecond * 2)

	found, _, _, _ = cache.Get(testKeyName) //nolint:dogsled
	assert.False(t, found)
}

func TestInMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewInMemoryCache(time.Minute, time.Microsecond)
	defer cache.Close()

	testCtx, testCancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case <-testCtx.Done():
				return
			default:
				_, _, _, _ = cache.Get("foo") //nolint:dogsled
			}
		}
	}()

	go func() {
		for {
			select {
			case <-testCtx.Done():
				return
			default:
				_, _ = cache.Delete("foo")
			}
		}
	}()

	go func() {
		for {
			select {
			case <-testCtx.Done():
				return
			default:
				_ = cache.Put("foo", []byte{1, 2, 3})
			}
		}
	}()

	<-time.After(time.Millisecond * 50)
	assert.NoError(t, cache.Close())

	<-time.After(time.Millisecond * 50)
	testCancel()
}

func TestInMemoryCache_Close(t *testing.T) {
	const testKeyName = "foo"

	cache := NewInMemoryCache(time.Millisecond*100, time.Millisecond)
	defer cache.Close()

	assert.NoError(t, cache.Put(testKeyName, []byte{1, 2, 3}))

	assert.NoError(t, cache.Close())

	<-time.After(time.Millisecond * 5)

	found, _, _, err := cache.Get(testKeyName)
	assert.False(t, found)
	assert.Equal(t, ErrClosed, err)

	err = cache.Put(testKeyName, []byte{1})
	assert.Equal(t, ErrClosed, err)

	ok, err := cache.Delete(testKeyName)
	assert.False(t, ok)
	assert.Equal(t, ErrClosed, err)

	err = cache.Close()
	assert.Equal(t, ErrClosed, err)
}

func TestInMemoryCache_GetWithEmptyKey(t *testing.T) {
	cache := NewInMemoryCache(time.Minute, time.Second)
	defer cache.Close()

	found, data, ttl, err := cache.Get("")
	assert.False(t, found)
	assert.Nil(t, data)
	assert.Zero(t, ttl)
	assert.Error(t, err)
}

func TestInMemoryCache_PutWithEmptyKey(t *testing.T) {
	cache := NewInMemoryCache(time.Minute, time.Second)
	defer cache.Close()

	err := cache.Put("", []byte{1})
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyKey, err)
}

func TestInMemoryCache_PutWithEmptyData(t *testing.T) {
	cache := NewInMemoryCache(time.Minute, time.Second)
	defer cache.Close()

	err := cache.Put("foo", []byte{})
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyData, err)

	err = cache.Put("foo", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyData, err)
}

func TestInMemoryCache_DeleteWithEmptyKey(t *testing.T) {
	cache := NewInMemoryCache(time.Minute, time.Second)
	defer cache.Close()

	deleted, err := cache.Delete("")
	assert.False(t, deleted)
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyKey, err)
}
