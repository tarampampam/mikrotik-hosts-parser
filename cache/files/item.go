package files

import (
	"io"
	"os"
	"sync"
	"time"
)

type hotBuffer struct {
	buf              []byte
	maxLen           int
	ttl              time.Duration
	cleaningDeferred bool
}

type Item struct {
	mutex     *sync.Mutex
	hotBuffer *hotBuffer
	key       string
	filePath  string
}

func newHotBuffer(maxLen int, ttl time.Duration) *hotBuffer {
	return &hotBuffer{
		buf:    make([]byte, 0),
		maxLen: maxLen,
		ttl:    ttl,
	}
}

// NewItem creates cache item.
// Maximum hot buffer length should be defined in bytes (set `0` to disable hot cache).
func NewItem(filePath, key string, hotBufLen int, hotBufTTL time.Duration) *Item {
	return &Item{
		mutex:     &sync.Mutex{},
		hotBuffer: newHotBuffer(hotBufLen, hotBufTTL),
		key:       key,
		filePath:  filePath,
	}
}

func (hb *hotBuffer) clean() {
	hb.buf = make([]byte, 0)
}

// GetKey returns the key for the current cache item
func (i *Item) GetKey() string {
	return i.key
}

// Get retrieves the value of the item from the cache associated with this object's key
func (i *Item) Get(to io.Writer) error {
	// deferred hot buffer cleaning
	if i.hotBuffer.ttl > 0 && !i.hotBuffer.cleaningDeferred {
		i.hotBuffer.cleaningDeferred = true
		defer func(hb *hotBuffer) {
			go func(hb *hotBuffer) {
				time.Sleep(hb.ttl)

				// make sure that deferring state was not changed
				if i.hotBuffer.cleaningDeferred {
					hb.clean()
					hb.cleaningDeferred = false
				}
			}(hb)
		}(i.hotBuffer)
	}

	// lock self and hot buffer for preventing concurrent buffer/content reading
	i.mutex.Lock()
	defer i.mutex.Unlock()

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
			i.hotBuffer.clean()
		}

		// write just read data into writer
		if _, err := to.Write(buf); err != nil {
			return newError(ErrBufferWriting, "cannot write into target buffer", err)
		}
	}

	return nil
}

// IsHit confirms if the cache item lookup resulted in a cache hit
func (i *Item) IsHit() bool {
	// lock hot buffer and self for preventing concurrent buffer/content reading
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

// Set the value represented by this cache item
func (i *Item) Set(from io.Reader) error {
	// lock self and hot buffer
	i.mutex.Lock()
	defer i.mutex.Unlock()

	// reset hot buffer deferred cleaning state
	i.hotBuffer.cleaningDeferred = false

	// make hot buf cleaning
	if len(i.hotBuffer.buf) != 0 {
		i.hotBuffer.clean()
	}

	// try to open file for writing
	file, err := os.OpenFile(i.filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
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
			i.hotBuffer.clean()
		}
	}

	return nil
}

// ExpiresAt sets the expiration time for this cache item
func (i *Item) ExpiresAt(when time.Time) error {
	panic("implement me")
}
