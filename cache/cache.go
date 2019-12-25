package cache

import (
	"io"
	"time"
)

// Main cache interfaces idea was looked at <https://www.php-fig.org/psr/psr-6/>

// Item defines an interface for interacting with objects inside a cache
type Item interface {
	// Returns the key for the current cache item.
	GetKey() string

	// Retrieves the value of the item from the cache associated with this object's key.
	Get(to io.Writer) error

	// Confirms if the cache item lookup resulted in a cache hit.
	IsHit() bool

	// Sets the value represented by this cache item.
	Set(from io.Reader) error

	// Indicates if cache item expiration time is exceeded. If expiration data was not set - error will be returned.
	IsExpired() (bool, error)

	// Returns the expiration time for this cache item. If expiration doesn't set - nil will be returned.
	ExpiresAt() *time.Time

	// Sets the expiration time for this cache item.
	SetExpiresAt(when time.Time) error
}

// Pool generates CacheItemInterface objects
type Pool interface {
	// Returns a Cache Item representing the specified key
	GetItem(key string) Item

	// Returns a map of cache items
	GetItems(keys []string) map[string]Item

	// Confirms if the cache contains specified cache item
	HasItem(key string) bool

	// Deletes all items in the pool
	Clear() (bool, error)

	// Removes the item from the pool
	DeleteItem(key string) (bool, error)

	// Removes multiple items from the pool
	DeleteItems(keys []string) (bool, error)

	// Persists a cache item immediately
	Save(item Item) (bool, error)
}
