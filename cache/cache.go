package cache

import (
	"io"
	"time"
)

// Item defines an interface for interacting with objects inside a cache
type Item interface {
	GetKey() string                 // Returns the key for the current cache item
	Get(to io.Writer) error         // Retrieves the value of the item from the cache associated with this object's key
	IsHit() bool                    // Confirms if the cache item lookup resulted in a cache hit
	Set(data []byte) error          // Sets the value represented by this cache item
	ExpiresAt(when time.Time) error // Sets the expiration time for this cache item
}

// Pool generates CacheItemInterface objects
type Pool interface {
	GetItem(key string) (Item, error)        // Returns a Cache Item representing the specified key
	GetItems(keys []string) ([]Item, error)  // Returns a slice set of cache items
	HasItem(key string) (bool, error)        // Confirms if the cache contains specified cache item
	Clear() (bool, error)                    // Deletes all items in the pool
	DeleteItem(key string) (bool, error)     // Removes the item from the pool
	DeleteItems(keys []string) (bool, error) // Removes multiple items from the pool
	Save(item Item) (bool, error)            // Persists a cache item immediately
}
