package cache

import (
	"time"
)

// Cacher is a byte-based cache with TTL.
type Cacher interface {
	// TTL returns current cache values time-to-live.
	TTL() time.Duration

	// Get value associated with the key from the storage.
	Get(key string) (found bool, data []byte, ttl time.Duration, err error)

	// Put value into the storage.
	Put(key string, data []byte) error

	// Delete value from the storage with passed key.
	Delete(key string) (bool, error)
}
