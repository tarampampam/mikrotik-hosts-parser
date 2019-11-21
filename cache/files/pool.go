package files

import (
	"io/ioutil"
	"mikrotik-hosts-parser/cache"
	"mikrotik-hosts-parser/cache/files/file"
	"os"
	"sync"
)

type Pool struct {
	cacheDirPath string
	mutex        *sync.Mutex
}

// NewPool creates new cache items pool
func NewPool(cacheDirPath string) *Pool {
	return &Pool{
		cacheDirPath: cacheDirPath,
		mutex:        &sync.Mutex{},
	}
}

// GetItem returns a Cache Item representing the specified key
func (p *Pool) GetItem(key string) cache.Item {
	return NewItem(p.cacheDirPath, key)
}

// GetItems returns a map of cache items
func (p Pool) GetItems(keys []string) map[string]cache.Item {
	res := make(map[string]cache.Item)

	for _, key := range keys {
		res[key] = p.GetItem(key)
	}

	return res
}

// HasItem confirms if the cache contains specified cache item
func (p Pool) HasItem(key string) bool {
	return p.GetItem(key).IsHit()
}

// Clear deletes all items in the pool
func (p Pool) Clear() (bool, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.clear()
}

func (p Pool) clear() (bool, error) {
	files, err := ioutil.ReadDir(p.cacheDirPath)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		cacheFile, err := file.OpenRead(f.Name(), DefaultItemFileSignature)

		// skip "wrong" or errored file
		if err != nil || cacheFile == nil {
			continue
		}
		// verify file signature
		matched, err := cacheFile.SignatureMatched()
		if closeErr := cacheFile.Close(); closeErr != nil {
			return false, closeErr
		}
		// if file signature is ok and we have no errors - remove the file
		if matched && err == nil {
			if rmErr := os.Remove(f.Name()); rmErr != nil {
				return false, rmErr
			}
		}
	}

	return false, nil
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
