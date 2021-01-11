package cache

import "time"

// Cacher is a cache engine.
type Cacher interface {
	// Get retrieves value for the key from the storage.
	Get(key string) (found bool, data []byte, ttl time.Duration, err error)

	// Put value into the storage.
	Put(key string, data []byte) error

	// Delete value from the storage with passed key.
	Delete(key string) (bool, error)
}
