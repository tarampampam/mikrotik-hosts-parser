package files

import (
	"io"
	"os"
	"sync"
	"time"
)

type hotBuffer struct {
	mutex            *sync.Mutex
	buf              []byte
	maxLen           int
	TTL              time.Duration
	cleaningDeferred bool
}

type item struct {
	mutex     *sync.Mutex
	hotBuffer *hotBuffer
	key       string
	filePath  string
}

func newHotBuffer(maxLen int, TTL time.Duration) *hotBuffer {
	return &hotBuffer{
		mutex:  &sync.Mutex{},
		buf:    make([]byte, 0),
		maxLen: maxLen,
		TTL:    TTL,
	}
}

// NewItem creates cache item.
// Maximum hot buffer length should be defined in bytes (set `0` to disable hot cache).
func NewItem(filePath, key string, hotBufLen int, hotBufTTL time.Duration) *item {
	return &item{
		mutex:     &sync.Mutex{},
		hotBuffer: newHotBuffer(hotBufLen, hotBufTTL),
		key:       key,
		filePath:  filePath,
	}
}

func (i *item) GetKey() string {
	return i.key
}

func (i *item) Get(to io.Writer) error {
	// lock hot buffer for preventing concurrent buffer/content reading
	i.hotBuffer.mutex.Lock()
	defer i.hotBuffer.mutex.Unlock()

	// deferred hot buffer cleaning
	if i.hotBuffer.TTL > 0 && !i.hotBuffer.cleaningDeferred {
		i.hotBuffer.cleaningDeferred = true
		defer func(hb *hotBuffer) {
			go func(hb *hotBuffer) {
				time.Sleep(hb.TTL)

				hb.mutex.Lock() // @todo: make check for concurrent mutex locking
				defer hb.mutex.Unlock()

				hb.buf = make([]byte, 0)
				hb.cleaningDeferred = false
			}(hb)
		}(i.hotBuffer)
	}

	// check for data existing in hot buffer
	if len(i.hotBuffer.buf) > 0 {
		// if data exists (buffer is not empty) - write data into writer from hot buffer directly (without file reading)
		if _, err := to.Write(i.hotBuffer.buf); err != nil {
			return err
		}
		return nil
	}

	// now we should lock self also
	i.mutex.Lock()
	defer i.mutex.Unlock()

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
		if n, err := file.Read(buf); err != nil {
			read += n // calculate total read size
			if err != io.EOF {
				return err
			}
			break
		}

		// if read size less then maximum hot buffer size - we append just read content into hot buffer
		if read <= i.hotBuffer.maxLen {
			i.hotBuffer.buf = append(i.hotBuffer.buf, buf...)
		} else if len(i.hotBuffer.buf) != 0 { // otherwise we should clean buffer
			i.hotBuffer.buf = i.hotBuffer.buf[:0]
		}

		// write just read data into writer
		if _, err := to.Write(buf); err != nil {
			return err
		}
	}

	return nil
}

func (i *item) IsHit() bool {
	// check for file exists
	if info, err := os.Stat(i.filePath); err == nil && info.Mode().IsRegular() {
		return true
	}

	return false
}

func (i *item) Set(data []byte) error {
	panic("implement me")
}

func (i *item) ExpiresAt(when time.Time) error {
	panic("implement me")
}
