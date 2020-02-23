package script

import (
	"github.com/tarampampam/go-filecache"
)

// defaultCachePool becomes accessible only after initial function calling
var defaultCachePool filecache.CachePool

// initDefaultCachePool makes default cache pool initialization
func initDefaultCachePool(cacheDir string, force bool) {
	if defaultCachePool == nil || !force {
		defaultCachePool = filecache.NewPool(cacheDir)
	}
}
