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
	return p.getItem(key)
}

func (p *Pool) getItem(key string) *Item {
	return NewItem(p.cacheDirPath, key)
}

// GetItems returns a map of cache items
func (p *Pool) GetItems(keys []string) map[string]cache.Item {
	res := make(map[string]cache.Item)

	for _, key := range keys {
		res[key] = p.getItem(key)
	}

	return res
}

// HasItem confirms if the cache contains specified cache item
func (p *Pool) HasItem(key string) bool {
	return p.getItem(key).IsHit()
}

func (p *Pool) walkCacheFiles(fn func(os.FileInfo)) error {
	files, err := ioutil.ReadDir(p.cacheDirPath)
	if err != nil {
		return err
	}

	for _, f := range files {
		cacheFile, err := file.OpenRead(f.Name(), DefaultItemFileSignature)

		// skip "wrong" or errored file
		if err != nil || cacheFile == nil {
			continue
		}
		// verify file signature and close file (closing error will be skipped)
		matched, _ := cacheFile.SignatureMatched()
		if closeErr := cacheFile.Close(); matched && closeErr == nil {
			// if all is ok - fall the func
			fn(f)
		}
	}

	return nil
}

// Clear deletes all items in the pool
func (p *Pool) Clear() (bool, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.clear()
}

func (p *Pool) clear() (bool, error) {
	var lastErr error

	err := p.walkCacheFiles(func(info os.FileInfo) {
		if rmErr := os.Remove(info.Name()); rmErr != nil {
			lastErr = rmErr
		}
	})

	if err != nil {
		return false, err
	}

	if lastErr != nil {
		return false, lastErr
	}

	return true, nil
}

// DeleteItem removes the item from the pool
func (p *Pool) DeleteItem(key string) (bool, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.deleteItem(key)
}

// DeleteItem removes the item from the pool
func (p *Pool) deleteItem(key string) (bool, error) {
	if rmErr := os.Remove(p.getItem(key).getFilePath()); rmErr != nil {
		return false, rmErr
	}

	return true, nil
}

// DeleteItems removes multiple items from the pool
func (p *Pool) DeleteItems(keys []string) (bool, error) {
	var lastErr error

	for _, key := range keys {
		if ok, delErr := p.deleteItem(key); !ok || delErr != nil {
			lastErr = delErr
		}
	}

	if lastErr != nil {
		return false, lastErr
	}

	return true, nil
}

func (p *Pool) Save(item cache.Item) (bool, error) {
	panic("implement me")
}
