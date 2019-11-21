package files

import (
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"mikrotik-hosts-parser/cache/files/file"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Item struct {
	mutex    *sync.Mutex
	hashing  hash.Hash
	dirPath  string
	fileName string
	key      string
}

// DefaultItemFilePerms is default permissions for file, associated with cache item
var DefaultItemFilePerms os.FileMode = 0664

// DefaultItemFileSignature is default signature for cache files
var DefaultItemFileSignature file.FSignature = nil

// NewItem creates cache item.
func NewItem(dirPath, key string) *Item {
	item := &Item{
		mutex:   &sync.Mutex{},
		hashing: sha1.New(), //nolint:gosec
		key:     key,
		dirPath: dirPath,
	}

	// generate file name based on hashed key value
	item.fileName = hex.EncodeToString(item.hashing.Sum([]byte(key)))

	return item
}

// GetKey returns the key for the current cache item.
func (i *Item) GetKey() string { return i.key }

// GetFilePath returns path to the associated file.
func (i *Item) getFilePath() string { return filepath.Join(i.dirPath, i.fileName) }

// Get retrieves the value of the item from the cache associated with this object's key.
func (i *Item) Get(to io.Writer) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.get(to)
}

func (i *Item) get(to io.Writer) error {
	// try to open file for reading
	f, openErr := file.Open(i.getFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if openErr != nil {
		return newError(ErrFileOpening, fmt.Sprintf("file [%s] cannot be opened", i.getFilePath()), openErr)
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.GetData(to); err != nil {
		return newError(ErrFileReading, fmt.Sprintf("file [%s] read error", i.getFilePath()), err)
	}

	return nil
}

// IsHit confirms if the cache item lookup resulted in a cache hit.
func (i *Item) IsHit() bool {
	i.mutex.Lock() // @todo: blocking is required?
	defer i.mutex.Unlock()

	return i.isHit()
}

func (i *Item) isHit() bool {
	// check for file exists
	if info, err := os.Stat(i.getFilePath()); err == nil && info.Mode().IsRegular() {
		return true
	}

	return false
}

// Set the value represented by this cache item.
func (i *Item) Set(from io.Reader) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.set(from)
}

// openOrCreateFile opens OR create file for item
func (i *Item) openOrCreateFile(filePath string, perm os.FileMode, signature file.FSignature) (*file.File, error) {
	if info, err := os.Stat(filePath); err == nil && info.Mode().IsRegular() {
		opened, openErr := file.Open(filePath, perm, signature)
		if openErr != nil {
			return nil, newError(ErrFileOpening, fmt.Sprintf("file [%s] cannot be opened", filePath), openErr)
		}
		return opened, nil
	}

	created, createErr := file.Create(filePath, perm, signature)
	if createErr != nil {
		return nil, newError(ErrFileWriting, fmt.Sprintf("cannot create file [%s]", filePath), createErr)
	}
	return created, nil
}

func (i *Item) set(from io.Reader) error {
	var filePath = i.getFilePath()

	f, err := i.openOrCreateFile(filePath, DefaultItemFilePerms, DefaultItemFileSignature)
	if err != nil {
		return err
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.SetData(from); err != nil {
		return newError(ErrFileWriting, fmt.Sprintf("cannot write into file [%s]", filePath), err)
	}

	return nil
}

// Indicates if cache item expiration time is exceeded. If expiration data was not set - error will be returned.
func (i *Item) IsExpired() (bool, error) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.isExpired()
}

func (i *Item) isExpired() (bool, error) {
	exp, expErr := i.expiresAt()

	if exp != nil {
		return exp.UnixNano() < time.Now().UnixNano(), nil
	}

	return false, newError(ErrExpirationDataNotAvailable, "expiration data reading error", expErr)
}

// ExpiresAt returns the expiration time for this cache item. If expiration doesn't set - nil will be returned.
// Important notice: returned time will be WITHOUT nanoseconds (just milliseconds).
func (i *Item) ExpiresAt() *time.Time {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	exp, _ := i.expiresAt()

	return exp
}

func (i *Item) expiresAt() (*time.Time, error) {
	f, openErr := file.Open(i.getFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if openErr != nil {
		return nil, openErr
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	exp, expErr := f.GetExpiresAt()

	if expErr != nil {
		return nil, expErr
	}

	return &exp, nil
}

// SetExpiresAt sets the expiration time for this cache item.
// Important notice: time will set WITHOUT nanoseconds (just milliseconds).
func (i *Item) SetExpiresAt(when time.Time) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.setExpiresAt(when)
}

func (i *Item) setExpiresAt(when time.Time) error {
	f, err := i.openOrCreateFile(i.getFilePath(), DefaultItemFilePerms, DefaultItemFileSignature)
	if err != nil {
		return err
	}
	defer func(f *file.File) { _ = f.Close() }(f)

	if err := f.SetExpiresAt(when); err != nil {
		return err
	}

	return nil
}
