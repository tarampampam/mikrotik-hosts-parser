package file

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"time"
)

// File offset type
type fOffset uint16

// File type
type FType string

const (
	// File block offsets are below:
	// +------------+-----------------------+--------------------+--------------+
	// | FType 0..7 |  FMeta 8..247         | FDataSHA1 248..288 | FData 289..n |
	// +------------+-----------------------+--------------------+--------------+
	// |            | ExpiresAtUnixMs 8..22 |                    |              |
	// +------------+-----------------------+--------------------+--------------+

	// File type bits (0..7)
	oFTypeFrom fOffset = 0
	oFTypeTo   fOffset = 7
	// File meta-information
	oFMetaFrom fOffset = oFTypeTo + 1
	// 14 bits for storing ExpiresAt in unix MILLI-seconds (1000 millisecond is equals to 1 second)
	oFMetaExpAtUnixMsFrom fOffset = oFMetaFrom
	oFMetaExpAtUnixMsTo   fOffset = oFMetaExpAtUnixMsFrom + 14
	// bits 22..246 are reserved
	oFMetaTo fOffset = 247
	// SHA1 hast of stored data
	oFDataSHA1From fOffset = oFMetaTo + 1
	oFDataSHA1To   fOffset = oFDataSHA1From + 40
	// Useful data
	oFDataFrom fOffset = oFDataSHA1To + 1
)

// File type definitions (important: max length is 8 bytes, UTF-8)
const (
	TRegularCacheEntry FType = "CACHE"
	TUnknown           FType = "UNKNOWN"
)

// Block type
type blockType byte

// Block type definitions
const (
	bFType            blockType = iota // File type
	bFMeta                             // Meta-information
	bFMetaExpAtUnixMS                  // Meta - ExpiresAt in unix milliseconds
	bFDataSHA1                         // Data SHA1 hash
)

// Cache file representation
type File struct {
	file *os.File
}

// Open opens the named file for reading. If successful, methods on the returned file can be used for reading; the
// associated file descriptor has mode O_RDONLY.
// If there is an error, it will be of type *os.PathError.
func Open(name string) (*File, error) {
	f, err := os.OpenFile(name, os.O_RDONLY, 0)
	return &File{file: f}, err
}

// Create creates or truncates the named file. If the file already exists, it is truncated. If the file does not exist,
// it is created with passed mode (permissions). If successful, methods on the returned File can be used for I/O; the
// associated file descriptor has mode O_RDWR.
// If there is an error, it will be of type *PathError.
func Create(name string, perm os.FileMode) (*File, error) {
	f, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	return &File{file: f}, err
}

// Name returns the name of the file as presented to Open.
func (f *File) Name() string { return f.file.Name() }

// Close closes the File, rendering it unusable for I/O. On files that support SetDeadline, any pending I/O operations
// will be canceled and return immediately with an error.
// Close will return an error if it has already been called.
func (f *File) Close() error {
	return f.file.Close()
}

// Remove "empty" (which is equals to zeros) bytes from slice.
func (f *File) rmEmptyBytes(in []byte) []byte {
	// calculate offset with "empty" bytes on lent side
	var off byte = 0
	for i, el := range in {
		if el != 0 {
			off = byte(i)
			break
		}
	}

	return in[off:]
}

// getBlockPosition returns offset for passed block type.
func (f *File) getBlockPosition(bType blockType) (from, to fOffset) {
	switch bType {
	case bFType:
		from = oFTypeFrom
		to = oFTypeTo
	case bFMeta:
		from = oFMetaFrom
		to = oFMetaTo
	case bFMetaExpAtUnixMS:
		from = oFMetaExpAtUnixMsFrom
		to = oFMetaExpAtUnixMsTo
	case bFDataSHA1:
		from = oFDataSHA1From
		to = oFDataSHA1To
	}
	return
}

// GetType of cache file. If file provided by this package - type should be `TRegularCacheEntry`.
func (f *File) GetType() (FType, error) {
	typeBytes, err := f.getType()

	switch FType(typeBytes) {
	case TRegularCacheEntry:
		return TRegularCacheEntry, nil
	default:
		return TUnknown, err
	}
}

// getType of a file as a bytes slice (with "empty" bytes removing), declared in file content.
func (f *File) getType() ([]byte, error) {
	from, to := f.getBlockPosition(bFType)
	buf := make([]byte, to-from+1)
	_, err := f.file.ReadAt(buf, int64(from))

	if err != nil && err != io.EOF {
		return nil, err
	}

	return f.rmEmptyBytes(buf), nil
}

// SetType of cache file. During working with this package you should set `TRegularCacheEntry` type.
func (f *File) SetType(fType FType) error {
	return f.setType(fType)
}

// setType of a cache file. This function converts type into bytes slice with "empty prefix" and write it into required
// position in a file.
func (f *File) setType(fType FType) error {
	from, to := f.getBlockPosition(bFType)
	buf := make([]byte, to-from+1)

	if len(fType) > len(buf) {
		return errors.New("cannot set file type: type is too long")
	}

	// fill-up bytes slice from end to start
	var j = byte(len(buf) - 1)
	for i := len(fType) - 1; i >= 0; i-- {
		buf[j] = fType[i]
		j--
	}

	n, err := f.file.WriteAt(buf, int64(from))

	if n != len(buf) {
		return errors.New("wrong wrote bytes length")
	}

	return err
}

// GetExpiresAt for current file (with milliseconds).
func (f *File) GetExpiresAt() (time.Time, error) {
	ms, err := f.getExpiresAtUnixMs()

	// check for "value was set?"
	if ms == 0 && err == nil {
		err = errors.New("value was not set")
	}

	return time.Unix(0, int64(ms*uint64(time.Millisecond))), err
}

// getExpiresAtUnixMs returns unsigned integer value with ExpiresAt in UNIX timestamp format in milliseconds.
func (f *File) getExpiresAtUnixMs() (uint64, error) {
	from, to := f.getBlockPosition(bFMetaExpAtUnixMS)
	buf := make([]byte, to-from)
	_, err := f.file.ReadAt(buf, int64(from))

	if err != nil && err != io.EOF {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf), nil
}

// SetExpiresAt sets the expiration time for current file.
func (f *File) SetExpiresAt(t time.Time) error {
	return f.setExpiresAtUnixMs(uint64(t.UnixNano() / int64(time.Millisecond)))
}

// setExpiresAtUnixMs takes unix timestamp (in milliseconds) and write them into required file block as a bytes slice.
func (f *File) setExpiresAtUnixMs(ts uint64) error {
	from, to := f.getBlockPosition(bFMetaExpAtUnixMS)
	buf := make([]byte, to-from)

	// pack unsigned integer into slice of bytes
	binary.LittleEndian.PutUint64(buf, ts)

	n, err := f.file.WriteAt(buf, int64(from))
	if n != len(buf) {
		return errors.New("wrong wrote bytes length")
	}

	return err
}
