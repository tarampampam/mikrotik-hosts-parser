package cache

import (
	"sync"
	"time"
)

type (
	// InMemoryCache is an inmemory cache (with TTL) implementation.
	InMemoryCache struct {
		ttl time.Duration
		ci  time.Duration // cleanup interval

		storageMu sync.RWMutex
		storage   map[string]inmemoryItem

		close    chan struct{}
		closedMu sync.RWMutex
		closed   bool
	}

	inmemoryItem struct {
		data          []byte
		expiresAtNano int64
	}
)

// NewInMemoryCache creates inmemory storage with TTL.
func NewInMemoryCache(ttl time.Duration, ci time.Duration) *InMemoryCache {
	cache := &InMemoryCache{ttl: ttl, ci: ci, storage: make(map[string]inmemoryItem), close: make(chan struct{}, 1)}
	go cache.cleanup()

	return cache
}

func (c *InMemoryCache) cleanup() {
	defer close(c.close)

	timer := time.NewTimer(c.ci)
	defer timer.Stop()

	for {
		select {
		case <-c.close:
			c.storageMu.Lock()
			for key := range c.storage {
				delete(c.storage, key)
			}
			c.storageMu.Unlock()

			return

		case <-timer.C:
			c.storageMu.Lock()
			var now = time.Now().UnixNano()

			for key, item := range c.storage {
				if now > item.expiresAtNano {
					delete(c.storage, key)
				}
			}
			c.storageMu.Unlock()

			timer.Reset(c.ci)
		}
	}
}

func (c *InMemoryCache) isClosed() bool {
	c.closedMu.RLock()
	defer c.closedMu.RUnlock()

	return c.closed
}

// Close current in memory storage with data invalidation.
func (c *InMemoryCache) Close() error {
	if c.isClosed() {
		return ErrClosed
	}

	c.closedMu.Lock()
	c.closed = true
	c.closedMu.Unlock()

	c.close <- struct{}{}

	return nil
}

// Get value associated with the key from the storage.
func (c *InMemoryCache) Get(key string) (bool, []byte, time.Duration, error) {
	if c.isClosed() {
		return false, nil, 0, ErrClosed
	}

	if key == "" {
		return false, nil, 0, ErrEmptyKey
	}

	c.storageMu.RLock()
	item, ok := c.storage[key]
	c.storageMu.RUnlock()

	if ok {
		now := time.Now()

		// item has been expired?
		if now.UnixNano() > item.expiresAtNano {
			c.storageMu.Lock()
			delete(c.storage, key)
			c.storageMu.Unlock()

			return false, nil, 0, nil
		}

		return true, item.data, time.Unix(0, item.expiresAtNano).Sub(now), nil
	}

	return false, nil, 0, nil
}

// Put value into the storage.
func (c *InMemoryCache) Put(key string, data []byte) error {
	if c.isClosed() {
		return ErrClosed
	}

	if key == "" {
		return ErrEmptyKey
	} else if len(data) == 0 {
		return ErrEmptyData
	}

	c.storageMu.Lock()
	c.storage[key] = inmemoryItem{data: data, expiresAtNano: time.Now().Add(c.ttl).UnixNano()}
	c.storageMu.Unlock()

	return nil
}

// Delete value from the storage with passed key.
func (c *InMemoryCache) Delete(key string) (bool, error) {
	if c.isClosed() {
		return false, ErrClosed
	}

	if key == "" {
		return false, ErrEmptyKey
	}

	c.storageMu.Lock()
	defer c.storageMu.Unlock()

	if _, ok := c.storage[key]; ok {
		delete(c.storage, key)

		return true, nil
	}

	return false, nil
}
