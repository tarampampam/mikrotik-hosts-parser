package file

import (
	"errors"
	"os"
)

// File offset type
type fOffset uint16

// File type
type FType string

const (
	// File block offsets are below:
	// +------------+-----------------+--------------------+--------------+
	// | FType 0..7 |  FMeta 8..247   | FDataSHA1 248..288 | FData 289..n |
	// +------------+-----------------+--------------------+--------------+
	// |            | TTLUnixMs 8..22 |                    |              |
	// +------------+-----------------+--------------------+--------------+

	// File type bits (0..7)
	oFTypeFrom fOffset = 0
	oFTypeTo   fOffset = 7
	// File meta-information
	oFMetaFrom fOffset = oFTypeTo + 1
	// 14 bits for storing TTL in unix MILLI-seconds (1000 millisecond is equals to 1 second)
	oFMetaTTLUnixMsFrom fOffset = oFMetaFrom
	oFMetaTTLUnixMsTo   fOffset = oFMetaTTLUnixMsFrom + 14
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
	tRegularCacheEntry FType = "CACHE"
	tUnknown           FType = "UNKNOWN"
)

// Block type
type blockType byte

// Block type definitions
const (
	bFType          blockType = iota // File type
	bFMeta                           // Meta-information
	bFMetaTTLUnixMS                  // Meta - TTL in unix milliseconds
	bFDataSHA1                       // Data SHA1 hash
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

// getBlockPosition returns offset for passed block type.
func (f *File) getBlockPosition(bType blockType) (from fOffset, to fOffset) {
	switch bType {
	case bFType:
		from = oFTypeFrom
		to = oFTypeTo
	case bFMeta:
		from = oFMetaFrom
		to = oFMetaTo
	case bFMetaTTLUnixMS:
		from = oFMetaTTLUnixMsFrom
		to = oFMetaTTLUnixMsTo
	case bFDataSHA1:
		from = oFDataSHA1From
		to = oFDataSHA1To
	}
	return
}

// GetType of cache file. If file provided by this package - type should be `tRegularCacheEntry`.
func (f *File) GetType() (FType, error) {
	typeBytes, err := f.getType()

	switch FType(typeBytes) {
	case tRegularCacheEntry:
		return tRegularCacheEntry, nil
	}

	return tUnknown, err
}

// getType of a file as a bytes slice (with "empty" bytes removing), declared in file content.
func (f *File) getType() ([]byte, error) {
	from, to := f.getBlockPosition(bFType)
	buf := make([]byte, to-from+1)
	n, err := f.file.ReadAt(buf, int64(from))

	if err != nil {
		return nil, err
	} else if n != len(buf) {
		return nil, errors.New("wrong bytes read length")
	}

	// calculate offset with "empty" bytes on lent side
	var off byte = 0
	for i, el := range buf {
		if el != 0 {
			off = byte(i)
			break
		}
	}

	return buf[off:], nil
}

// SetType of cache file. During working with this package you should set `tRegularCacheEntry` type.
func (f *File) SetType(fType FType) error {
	return f.setType(fType)
}

// setType of a cache file. This function converts type into bytes slice with "empty prefix" and write it into required
// position in a file.
func (f *File) setType(fType FType) error {
	if len(fType) > 8 {
		return errors.New("cannot set file type: type is too long")
	}

	from, to := f.getBlockPosition(bFType)
	buf := make([]byte, to-from+1)

	// fill-up bytes slice from end to start
	var j = byte(len(buf) - 1)
	for i := len(fType) - 1; i >= 0; i-- {
		buf[j] = fType[i]
		j--
	}

	n, err := f.file.WriteAt(buf, int64(from))

	if n != len(buf) {
		return errors.New("wrong bytes read length")
	}

	return err
}
