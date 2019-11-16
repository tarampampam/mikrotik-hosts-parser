package files

import (
	"encoding/binary"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

type hotBuffer struct {
	buf                 []byte
	exp                 *time.Time
	maxLen              int
	ttl                 time.Duration
	bufCleaningDeferred bool
	expCleaningDeferred bool
}

type Item struct {
	mutex     *sync.Mutex
	hotBuffer *hotBuffer
	key       string
	filePath  string
	perm      os.FileMode
}

// NewItem creates cache item.
// Maximum hot buffer length should be defined in bytes (set `0` to disable hot cache).
func NewItem(filePath, key string, hotBufLen int, hotBufTTL time.Duration) *Item {
	return &Item{
		mutex: &sync.Mutex{},
		hotBuffer: &hotBuffer{
			buf:    make([]byte, 0),
			exp:    nil,
			maxLen: hotBufLen,
			ttl:    hotBufTTL,
		},
		key:      key,
		filePath: filePath,
		perm:     0664, // creation files permissions
	}
}

// Make buffer cleaning.
func (hb *hotBuffer) bufClean() {
	hb.buf = make([]byte, 0)
}

// GetKey returns the key for the current cache item.
func (i *Item) GetKey() string {
	return i.key
}

// expFilePath returns path to the file with expiration timestamp.
func (i *Item) expFilePath() string {
	return i.filePath + ".expire"
}

// deferred hot buffer cleaning (if needed).
func (i *Item) deferCleanHotBuf() {
	i.hotBuffer.bufCleaningDeferred = true
	go func(i *Item) {
		time.Sleep(i.hotBuffer.ttl)

		i.mutex.Lock()
		defer i.mutex.Unlock()

		// make sure that deferring state was not changed
		if i.hotBuffer.bufCleaningDeferred {
			i.hotBuffer.bufClean()
			i.hotBuffer.bufCleaningDeferred = false
		}
	}(i)
}

// Get retrieves the value of the item from the cache associated with this object's key.
func (i *Item) Get(to io.Writer) error {
	// lock self for preventing concurrent buffer/content reading
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// deferred hot buffer cleaning
	if i.hotBuffer.ttl > 0 && !i.hotBuffer.bufCleaningDeferred {
		defer i.deferCleanHotBuf()
	}

	// check for data existing in hot buffer
	if len(i.hotBuffer.buf) > 0 {
		// if data exists (buffer is not empty) - write data into writer from hot buffer directly (without file reading)
		if _, err := to.Write(i.hotBuffer.buf); err != nil {
			return newError(ErrBufferWriting, "cannot write into target buffer", err)
		}
		return nil
	}

	// try to open file for reading
	file, err := os.Open(i.filePath)
	if err != nil {
		return newErrorf(ErrFileOpening, err, "file [%s] cannot be opened", i.filePath)
	}
	defer file.Close()

	// read file using buffer
	buf := make([]byte, 32)
	read := 0
	for {
		// read part of file
		readBytes, err := file.Read(buf)
		if err != nil {
			if err != io.EOF {
				return newErrorf(ErrFileReading, err, "file [%s] cannot be read", i.filePath)
			}
			break
		}

		// calculate total read size
		read += readBytes

		// limit buffer size to actual read bytes length
		if readBytes != len(buf) {
			buf = buf[0:readBytes]
		}

		// if read size less then maximum hot buffer size - we append just read content into hot buffer
		if read <= i.hotBuffer.maxLen {
			i.hotBuffer.buf = append(i.hotBuffer.buf, buf...)
		} else if len(i.hotBuffer.buf) != 0 { // otherwise we should clean buffer
			i.hotBuffer.bufClean()
		}

		// write just read data into writer
		if _, err := to.Write(buf); err != nil {
			return newError(ErrBufferWriting, "cannot write into target buffer", err)
		}
	}

	return nil
}

// IsHit confirms if the cache item lookup resulted in a cache hit.
func (i *Item) IsHit() bool {
	// lock self for preventing concurrent buffer/content reading
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// fast check based on hot buffer size - if size is not equals zero - file must be exists
	if len(i.hotBuffer.buf) != 0 {
		return true
	}

	// check for file exists
	if info, err := os.Stat(i.filePath); err == nil && info.Mode().IsRegular() {
		return true
	}

	return false
}

// Set the value represented by this cache item.
func (i *Item) Set(from io.Reader) error {
	// lock self for preventing concurrent buffer/content reading
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// reset hot buffer deferred cleaning state
	i.hotBuffer.bufCleaningDeferred = false

	// make hot buf cleaning
	if len(i.hotBuffer.buf) != 0 {
		i.hotBuffer.bufClean()
	}

	// try to open file for writing
	file, err := os.OpenFile(i.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, i.perm)
	if err != nil {
		return newErrorf(ErrFileOpening, err, "file [%s] cannot be opened", i.filePath)
	}
	defer file.Close()

	// read file using buffer
	buf := make([]byte, 32)
	wrote := 0
	for {
		// read part of input data
		readBytes, err := from.Read(buf)
		if err != nil {
			if err != io.EOF {
				return newError(ErrFileReading, "source buffer cannot be read", err)
			}
			break
		}

		// limit buffer size to actual read bytes length
		if readBytes != len(buf) {
			buf = buf[0:readBytes]
		}

		// write this part into file
		n, err := file.Write(buf)
		wrote += n
		if err != nil {
			return newErrorf(ErrFileWriting, err, "cannot write into file [%s]", file.Name())
		}

		// and if read size less then maximum hot buffer size - we append just read content into hot buffer
		if wrote <= i.hotBuffer.maxLen {
			i.hotBuffer.buf = append(i.hotBuffer.buf, buf...)
		} else if len(i.hotBuffer.buf) != 0 { // otherwise we should clean buffer
			i.hotBuffer.bufClean()
		}
	}

	// deferred hot buffer cleaning
	if len(i.hotBuffer.buf) != 0 && i.hotBuffer.ttl > 0 && !i.hotBuffer.bufCleaningDeferred {
		defer i.deferCleanHotBuf()
	}

	return nil
}

// delayed cleaning expiring data in hot buffer.
func (i *Item) deferCleanHotBufExpData() {
	i.hotBuffer.expCleaningDeferred = true
	go func(i *Item) {
		time.Sleep(i.hotBuffer.ttl)

		i.mutex.Lock()
		defer i.mutex.Unlock()

		// make sure that deferring state was not changed
		if i.hotBuffer.expCleaningDeferred {
			i.hotBuffer.exp = nil
			i.hotBuffer.expCleaningDeferred = false
		}
	}(i)
}

// Indicates if cache item expiration time is exceeded. If expiration data was not set - error will be returned.
func (i *Item) IsExpired() (bool, error) {
	if exp := i.ExpiresAt(); exp != nil {
		return exp.UnixNano() < time.Now().UnixNano(), nil
	}

	return false, newError(ErrExpirationDataNotAvailable, "expiration data is not available", nil)
}

// ExpiresAt returns the expiration time for this cache item. If expiration doesn't set - nil will be returned.
func (i *Item) ExpiresAt() *time.Time {
	// lock self for preventing concurrent buffer/content reading
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// defer expiring data cleaning (if needed)
	if i.hotBuffer.ttl > 0 && !i.hotBuffer.expCleaningDeferred {
		defer i.deferCleanHotBufExpData()
	}

	// check exp data in hot buffer at first
	if i.hotBuffer.exp != nil {
		return i.hotBuffer.exp
	}

	filePath := i.expFilePath()

	if file, err := os.Open(filePath); err == nil {
		defer file.Close()
		if data, err := ioutil.ReadAll(file); err == nil {
			// convert just read bytes slice into time struct
			exp := time.Unix(0, int64(binary.LittleEndian.Uint64(data)))

			// refresh exp time in hot buffer
			if !i.hotBuffer.expCleaningDeferred {
				i.hotBuffer.exp = &exp
			}

			return &exp
		}
	}

	return nil
}

// SetExpiresAt sets the expiration time for this cache item.
func (i *Item) SetExpiresAt(when time.Time) error {
	slice := make([]byte, 8)
	filePath := i.expFilePath()
	// make copy - `exp == when` but `&exp != &when`
	exp := when

	binary.LittleEndian.PutUint64(slice, uint64(exp.UnixNano()))

	// lock self for preventing concurrent buffer/content reading
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if err := ioutil.WriteFile(filePath, slice, i.perm); err != nil {
		return newErrorf(ErrFileWriting, err, "cannot write into file [%s]", filePath)
	}

	// set expiring data in hot buffer
	i.hotBuffer.exp = &exp

	// defer expiring data cleaning (if needed)
	if i.hotBuffer.ttl > 0 && !i.hotBuffer.expCleaningDeferred {
		defer i.deferCleanHotBufExpData()
	}

	return nil
}
