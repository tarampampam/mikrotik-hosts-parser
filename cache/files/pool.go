package files

import (
	"crypto/md5" // nolint:gosec
	"encoding/hex"
	"hash"
	"mikrotik-hosts-parser/cache"
	"strings"
	"sync"
	"time"
)

type Pool struct {
	cachePath string
	hotBufLen int
	hotBufTTL time.Duration
	mutex     *sync.Mutex
	items     map[string]cache.Item // "cache" of item objects
	hash      hash.Hash
}

func NewPool(cachePath string, hotBufLen int, hotBufTTL time.Duration) *Pool {
	return &Pool{
		cachePath: cachePath,
		hotBufLen: hotBufLen,
		hotBufTTL: hotBufTTL,
		mutex:     &sync.Mutex{},
		items:     make(map[string]cache.Item),
		hash:      md5.New(), // nolint:gosec // selected hash algorithm
	}
}

func (p *Pool) keyToHash(key string) string {
	slice := md5.Sum([]byte(key)) // nolint:gosec

	return strings.ToLower(hex.EncodeToString(slice[:]))
}

func (p *Pool) GetItem(key string) (cache.Item, error) {
	// hash passed key
	key = p.keyToHash(key)

	// check for existing in items map
	if item, ok := p.items[key]; ok {
		return item, nil
	}

	panic("implement me")
}

func (p Pool) GetItems(keys []string) ([]cache.Item, error) {
	panic("implement me")
}

func (p Pool) HasItem(key string) (bool, error) {
	panic("implement me")
}

func (p Pool) Clear() (bool, error) {
	panic("implement me")
}

func (p Pool) DeleteItem(key string) (bool, error) {
	panic("implement me")
}

func (p Pool) DeleteItems(keys []string) (bool, error) {
	panic("implement me")
}

func (p Pool) Save(item cache.Item) (bool, error) {
	panic("implement me")
}
