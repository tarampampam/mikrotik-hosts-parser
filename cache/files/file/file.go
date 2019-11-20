package file

import (
	"bytes"
	"crypto/sha1" //nolint:gosec
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"time"
)

// Read/write buffer size in bytes
const rwBufferSize byte = 32

type (
	// File signature
	FSignature []byte

	// File field offset and length
	offset uint8
	length uint8

	// File field for storing "File signature" (special flag for file identification among other files)
	ffSignature struct {
		offset
		length
	}

	// File field for storing "Expires At" label (in unix timestamp format with milliseconds)
	ffExpiresAtUnixMs struct {
		offset
		length
	}

	// File field for storing data "hash sum" (in SHA1 format)
	ffDataSha1 struct {
		offset
		length
	}

	// Field for useful data
	ffData struct {
		offset
	}

	// Cache file representation (all offsets must be set manually on instance creation action)
	File struct {
		ffSignature
		ffExpiresAtUnixMs
		ffDataSha1
		ffData
		Signature FSignature
		file      *os.File  // file on filesystem
		hashing   hash.Hash // SHA1 "generator" (required for hash sum calculation)
	}
)

var DefaultSignature = FSignature("#/CACHE ") // 35, 47, 67, 65, 67, 72, 69, 32

// Create new file instance.
func newFile(file *os.File, signature FSignature) *File {
	// setup default file type bytes slice
	if signature == nil || len(signature) == 0 {
		signature = DefaultSignature
	}

	// File block offsets are below:
	// +----------------+-----------------------+-----------------+------------+
	// | Signature 0..7 |    Meta Data 8..63    | DataSHA1 64..83 | Data 84..n |
	// +----------------+-----------------------+-----------------+------------+
	// |                | ExpiresAtUnixMs 8..22 |                 |            |
	// +----------------+-----------------------+-----------------+------------+
	// |                |    RESERVED 23..63    |                 |            |
	// +----------------+-----------------------+-----------------+------------+
	return &File{
		ffSignature: ffSignature{
			offset: 0,
			length: 8,
		},
		ffExpiresAtUnixMs: ffExpiresAtUnixMs{
			offset: 8,
			length: 14,
		},
		ffDataSha1: ffDataSha1{
			offset: 64,
			length: 20,
		},
		ffData: ffData{
			offset: 84,
		},
		Signature: signature,
		file:      file,
		hashing:   sha1.New(), //nolint:gosec
	}
}

// Create creates or truncates the named file. If the file already exists, it is truncated. If the file does not exist,
// it is created with passed mode (permissions).
// signature can be omitted (nil) - in this case will be used default file signature.
// Important: file with signature will be created immediately.
func Create(name string, perm os.FileMode, signature FSignature) (*File, error) {
	f, openErr := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if openErr != nil {
		return nil, openErr
	}

	file := newFile(f, signature)

	// write file signature
	if err := file.SetSignature(file.Signature); err != nil {
		return nil, err
	}

	if err := file.SetData(bytes.NewBuffer([]byte{})); err != nil {
		return nil, err
	}

	return file, nil
}

// Open opens the named file for reading. If successful, methods on the returned file can be used for reading; the
// associated file descriptor has mode O_RDONLY.
// If there is an error, it will be of type *os.PathError.
// signature can be omitted (nil) - in this case will be used default file signature.
func Open(name string, perm os.FileMode, signature FSignature) (*File, error) {
	f, err := os.OpenFile(name, os.O_RDWR, perm)
	if err != nil {
		return nil, err
	}

	return newFile(f, signature), nil
}

// Name returns the name of the file as presented to Open.
func (f *File) Name() string { return f.file.Name() }

// Close closes the File, rendering it unusable for I/O. On files that support SetDeadline, any pending I/O operations
// will be canceled and return immediately with an error.
// Close will return an error if it has already been called.
func (f *File) Close() error {
	return f.file.Close()
}

// SignatureMatched checks for file signature matching. Signature should be set on file creation. This function can
// helps you to detect files that created by current package.
func (f *File) SignatureMatched() (bool, error) {
	fType, err := f.getSignature()
	if err != nil {
		return false, err
	}

	return bytes.Equal(*fType, f.Signature), nil
}

// GetSignature of current file signature as a typed slice of a bytes.
func (f *File) GetSignature() (*FSignature, error) {
	return f.getSignature()
}

// getSignature of current file signature as a typed slice of a bytes.
func (f *File) getSignature() (*FSignature, error) {
	buf := make(FSignature, f.ffSignature.length)

	if n, err := f.file.ReadAt(buf, int64(f.ffSignature.offset)); err != nil && err != io.EOF {
		return nil, err
	} else if l := len(buf); n != l {
		// limit length for too small reading results
		buf = buf[0:n]
	}

	return &buf, nil
}

// SetSignature for current file.
func (f *File) SetSignature(signature FSignature) error {
	return f.setSignature(signature)
}

// setSignature allows to use only bytes slice of signature with length defined in file structure.
func (f *File) setSignature(signature FSignature) error {
	if l := len(signature); l != int(f.ffSignature.length) {
		return fmt.Errorf("wrong signature length: required length: %d, passed: %d", f.ffSignature.length, l)
	}

	if n, err := f.file.WriteAt(signature, int64(f.ffSignature.offset)); err != nil {
		return err
	} else if n != len(signature) {
		return errors.New("wrong wrote bytes length")
	}

	return nil
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
	buf := make([]byte, f.ffExpiresAtUnixMs.length)

	if _, err := f.file.ReadAt(buf, int64(f.ffExpiresAtUnixMs.offset)); err != nil && err != io.EOF {
		return 0, err
	}

	return binary.LittleEndian.Uint64(buf), nil
}

func (f *File) SetExpiresAt(t time.Time) error {
	return f.setExpiresAtUnixMs(uint64(t.UnixNano() / int64(time.Millisecond)))
}

func (f *File) setExpiresAtUnixMs(ts uint64) error {
	buf := make([]byte, f.ffExpiresAtUnixMs.length)

	// pack unsigned integer into slice of bytes
	binary.LittleEndian.PutUint64(buf, ts)

	if n, err := f.file.WriteAt(buf, int64(f.ffExpiresAtUnixMs.offset)); err != nil {
		return err
	} else if n != len(buf) {
		return errors.New("wrong wrote bytes length")
	}

	return nil
}

func (f *File) setDataSHA1(h []byte) error {
	if l := len(h); l != int(f.ffDataSha1.length) {
		return fmt.Errorf("wrong hash length: required length: %d, passed: %d", f.ffDataSha1.length, l)
	}

	if n, err := f.file.WriteAt(h, int64(f.ffDataSha1.offset)); err != nil {
		return err
	} else if n != len(h) {
		return errors.New("wrong wrote bytes length")
	}

	return nil
}

func (f *File) GetDataHash() ([]byte, error) {
	return f.getDataSHA1()
}

func (f *File) getDataSHA1() ([]byte, error) {
	buf := make([]byte, f.ffDataSha1.length)

	if _, err := f.file.ReadAt(buf, int64(f.ffDataSha1.offset)); err != nil && err != io.EOF {
		return buf, err
	}

	return buf, nil
}

func (f *File) SetData(in io.Reader) error {
	return f.setData(in)
}

func (f *File) setData(in io.Reader) error {
	buf := make([]byte, rwBufferSize)
	off := int64(f.ffData.offset)
	f.hashing.Reset()

	for {
		// read part of input data
		n, err := in.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}

		// limit length for too small reading results
		if l := len(buf); n != l {
			buf = buf[0:n]
		}

		// write content into required position
		wroteBytes, writeErr := f.file.WriteAt(buf, off)
		if writeErr != nil {
			return writeErr
		}
		// write into "hashing" too for hash sum calculation
		if _, err := f.hashing.Write(buf); err != nil {
			return err
		}

		// move offset
		off += int64(wroteBytes)
	}

	if err := f.setDataSHA1(f.hashing.Sum(nil)); err != nil {
		return err
	}

	return nil
}

func (f *File) GetData(out io.Writer) error {
	return f.getData(out)
}

func (f *File) getData(out io.Writer) error {
	buf := make([]byte, rwBufferSize)
	off := uint64(f.ffData.offset)
	f.hashing.Reset()

	for {
		// read part of useful data
		n, readErr := f.file.ReadAt(buf, int64(off))
		if readErr != nil {
			if readErr != io.EOF {
				return readErr
			}
		}
		// limit length for too small reading results
		if l := len(buf); n != l {
			buf = buf[0:n]
		}

		// write content into out writer
		wroteBytes, writeErr := out.Write(buf)
		if writeErr != nil {
			return writeErr
		}
		// write into "hashing" too for hash sum calculation
		if _, err := f.hashing.Write(buf); err != nil {
			return err
		}

		// move offset
		off += uint64(wroteBytes)

		if readErr != nil {
			break
		}
	}

	// calculate just read data hash
	dataHash := f.hashing.Sum(nil)

	// get existing hash
	existsHash, hashErr := f.getDataSHA1()
	if hashErr != nil {
		return hashErr
	}

	// if hashes mismatched - data was broken
	if !bytes.Equal(dataHash, existsHash) {
		return fmt.Errorf("data hashes mismatched. required: %v, current: %v", existsHash, dataHash)
	}

	return nil
}
