package file

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f, createErr := Create(filepath.Join(tmpDir, "a"), 0664, nil)
	if createErr != nil {
		t.Fatalf("Got unexpected error on file creation: %v", createErr)
	}
	defer func() { _ = f.Close() }()

	if !bytes.Equal(f.Signature, DefaultSignature) {
		t.Errorf("Created file has non-default signature. Got: %v, want: %v", f.Signature, DefaultSignature)
	}

	if ok, err := f.SignatureMatched(); !ok {
		t.Error("For just created file we'v got signature mismatch")
	} else if err != nil {
		t.Errorf("Got unexpected error: %v", err)
	}

	if info, err := os.Stat(f.Name()); os.IsNotExist(err) || info.IsDir() {
		t.Errorf("Required file [%s] does not exists", f.Name())
	}
}

func TestCreateUsingDifferentParameters(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	tests := []struct {
		name           string
		giveFilePath   string
		giveSignature  *FSignature
		givePerms      os.FileMode
		wantError      bool
		wantErrorWords []string
	}{
		{
			name:          "correct short signature",
			giveFilePath:  filepath.Join(tmpDir, "a"),
			giveSignature: &FSignature{0, 1, 2, 3, 4, 5, 6, 7},
			givePerms:     0664,
			wantError:     false,
		},
		{
			name:           "too short signature",
			giveFilePath:   filepath.Join(tmpDir, "b"),
			giveSignature:  &FSignature{0, 1, 2},
			givePerms:      0664,
			wantError:      true,
			wantErrorWords: []string{"wrong signature length", "required length: 8"},
		},
		{
			name:           "too long signature",
			giveFilePath:   filepath.Join(tmpDir, "c"),
			giveSignature:  &FSignature{0, 1, 2, 3, 4, 5, 6, 7, 8},
			givePerms:      0664,
			wantError:      true,
			wantErrorWords: []string{"wrong signature length", "required length: 8"},
		},
		{
			name:           "wrong file destination",
			giveFilePath:   tmpDir,
			giveSignature:  &DefaultSignature,
			givePerms:      0664,
			wantError:      true,
			wantErrorWords: []string{"is a directory", tmpDir},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, createErr := Create(tt.giveFilePath, tt.givePerms, *tt.giveSignature)

			if createErr == nil && tt.wantError {
				t.Fatal("Expected error was not returned")
			} else if createErr != nil && !tt.wantError {
				t.Fatalf("Got unexpected error on file creation: %v", createErr)
			} else if createErr != nil && tt.wantError {
				for _, word := range tt.wantErrorWords {
					if !strings.Contains(createErr.Error(), word) {
						t.Errorf("Required word [%s] was not found in eror message [%s]", word, createErr.Error())
					}
				}
			}

			if f != nil {
				_ = f.Close()
			}
		})
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f1, _ := Create(filepath.Join(tmpDir, "a"), 0664, nil)
	f2, _ := os.Create(filepath.Join(tmpDir, "b"))
	_, _ = f2.WriteString("foo bar")
	_ = f1.Close()
	_ = f2.Close()

	file1, file1err := Open(f1.Name(), 0664, nil)
	if file1err != nil {
		t.Errorf("Got unexpected error: %v", file1err)
	}
	if ok, _ := file1.SignatureMatched(); !ok {
		t.Error("Signature mismatched for correct file")
	}
	_ = file1.Close()

	file2, file2err := Open(f2.Name(), 0664, nil)
	if file2err != nil {
		t.Errorf("Got unexpected error: %v", file1err)
	}
	if ok, _ := file2.SignatureMatched(); ok {
		t.Error("Signature must be mismatched for incorrect file")
	}
	_ = file2.Close()

	_, file3err := Open(tmpDir, 0664, nil)
	if file3err == nil {
		t.Error("Expected error not returned")
	}
}

func TestFile_Name(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	name := filepath.Join(tmpDir, "a")
	f, _ := Create(name, 0664, nil)

	if f.Name() != name {
		t.Errorf("Wrong name. Want [%s], got: [%s]", name, f.Name())
	}

	_ = f.Close()
}

func TestFile_GetAndSetExpiresAt(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f, _ := Create(filepath.Join(tmpDir, "a"), 0664, nil)
	defer func() { _ = f.Close() }()

	if v, err := f.GetExpiresAt(); err == nil {
		t.Errorf("Expected error was not returned (value is %v)", v)
	}

	exp := time.Now()

	if err := f.SetExpiresAt(exp); err != nil {
		t.Errorf("Got unexpected error on experation setting: %v", err)
	}

	gotExp, getErr := f.GetExpiresAt()
	if getErr != nil {
		t.Errorf("Got unexpected error on experation getting: %v", getErr)
	}

	if gotExp.Unix() != exp.Unix() {
		t.Errorf("Got wrong experation time value. Want: %v, got: %v", exp, gotExp)
	}
}

func TestFile_GetAndSetData(t *testing.T) { //nolint:gocyclo
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	// genRandomContent generates slice of random bytes with passed length
	genRandomContent := func(t *testing.T, len int) []byte {
		t.Helper()

		buf := make([]byte, 0)
		rand.Seed(time.Now().UnixNano())

		for i := 1; i <= len; i++ {
			buf = append(buf, byte(rand.Intn(255)))
		}

		return buf
	}

	tests := []struct {
		name        string
		giveFile    func(t *testing.T) *File // file fabric
		giveContent []byte
	}{
		{
			name: "without content",
			giveFile: func(t *testing.T) *File {
				f, err := Create(filepath.Join(tmpDir, "a"), 0664, nil)
				if err != nil {
					t.Fatalf("Got unexpected error on file creation: %v", err)
				}
				return f
			},
			giveContent: genRandomContent(t, 0),
		},
		{
			name: "8 bytes content",
			giveFile: func(t *testing.T) *File {
				f, err := Create(filepath.Join(tmpDir, "b"), 0664, nil)
				if err != nil {
					t.Fatalf("Got unexpected error on file creation: %v", err)
				}
				return f
			},
			giveContent: genRandomContent(t, 8),
		},
		{
			name: "medium content size",
			giveFile: func(t *testing.T) *File {
				f, err := Create(filepath.Join(tmpDir, "c"), 0664, nil)
				if err != nil {
					t.Fatalf("Got unexpected error on file creation: %v", err)
				}
				return f
			},
			giveContent: genRandomContent(t, 1024*2),
		},
		{
			name: "large content size",
			giveFile: func(t *testing.T) *File {
				f, err := Create(filepath.Join(tmpDir, "d"), 0664, nil)
				if err != nil {
					t.Fatalf("Got unexpected error on file creation: %v", err)
				}
				return f
			},
			giveContent: genRandomContent(t, 1024*512),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := tt.giveFile(t)
			defer func(f *File) { _ = f.Close() }(f)

			justCreatedHash, getHashErr := f.GetDataHash()
			// get hashsum
			if len(justCreatedHash) == 0 {
				t.Errorf("Hashsum for just created file should be non-empty: %v", justCreatedHash)
			} else if getHashErr != nil {
				t.Errorf("Got unexpected error on hashsum getting: %v", getHashErr)
			}

			exp := time.Now()

			// set expiration date/time
			if err := f.SetExpiresAt(exp); err != nil {
				t.Errorf("Got unexpected error on experation setting: %v", err)
			}

			// get data (should be empty)
			writeTo := bytes.NewBuffer([]byte{})
			if err := f.GetData(writeTo); err != nil {
				t.Errorf("Got unexpected error on data getting: %v", err)
			}
			if data := writeTo.Bytes(); len(data) != 0 {
				t.Errorf("For just created file data should be empty. Got: %v", data)
			}

			// set some data
			readFrom := bytes.NewBuffer(tt.giveContent)
			if err := f.SetData(readFrom); err != nil {
				t.Errorf("Got unexpected error on data setting: %v", err)
			}

			// hashsum must changes
			if len(tt.giveContent) > 0 {
				if h, err := f.GetDataHash(); bytes.Equal(justCreatedHash, h) {
					t.Errorf("Hashes after data setting [%v] and just created file [%v] should be different", h, justCreatedHash)
				} else if err != nil {
					t.Errorf("Got unexpected error on hashsum getting: %v", err)
				}
			}

			// get data back
			writeBackTo := bytes.NewBuffer([]byte{})
			if err := f.GetData(writeBackTo); err != nil {
				t.Errorf("Got unexpected error on data getting: %v", err)
			}
			if data := writeBackTo.Bytes(); !bytes.Equal(data, tt.giveContent) {
				t.Errorf("Wrong content returned. Want: %v, got: %v", tt.giveContent, data)
			}

			// check expiring date/time
			if fExp, err := f.GetExpiresAt(); fExp.Unix() != exp.Unix() {
				t.Errorf("Got unexpected expiring date/time. Want: %v, got: %v", exp, fExp)
			} else if err != nil {
				t.Errorf("Got unexpected error on experation getting: %v", err)
			}

			// for debug: `t.Log(ioutil.ReadAll(f.file))`
		})
	}
}

// Create temporary directory.
func createTempDir(t *testing.T) string {
	t.Helper()

	tmpDir, err := ioutil.TempDir("", "test-")
	if err != nil {
		t.Fatal(err)
	}

	return tmpDir
}

// Remove temporary directory.
func removeTempDir(t *testing.T, dirPath string) {
	t.Helper()

	if !strings.HasPrefix(dirPath, os.TempDir()) {
		t.Fatalf("Wrong tmp dir path: %s", dirPath)
		return
	}

	if err := os.RemoveAll(dirPath); err != nil {
		t.Fatal(err)
	}
}
