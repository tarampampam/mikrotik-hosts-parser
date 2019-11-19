package file

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	defer f.Close()

	if createErr != nil {
		t.Fatalf("Got unexpected error on file creation: %v", createErr)
	}

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
			name:           "correct short signature",
			giveFilePath:   filepath.Join(tmpDir, "a"),
			giveSignature:  &FSignature{0, 1, 2, 3, 4, 5, 6, 7},
			givePerms:      0664,
			wantError:      false,
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
			defer func() {
				if f != nil {
					f.Close()
				}
			}()

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
		})
	}
}

func TestOpen(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f1, _ := Create(filepath.Join(tmpDir, "a"), 0664, nil)
	f2, _ := os.Create(filepath.Join(tmpDir, "b"))
	f2.WriteString("foo bar")
	f1.Close()
	f2.Close()

	file1, file1err := Open(f1.Name(), 0664, nil)
	if file1err != nil {
		t.Errorf("Got unexpected error: %v", file1err)
	}
	if ok, _ := file1.SignatureMatched(); !ok {
		t.Error("Signature mismatched for correct file")
	}
	file1.Close()

	file2, file2err := Open(f2.Name(), 0664, nil)
	if file2err != nil {
		t.Errorf("Got unexpected error: %v", file1err)
	}
	if ok, _ := file2.SignatureMatched(); ok {
		t.Error("Signature must be mismatched for incorrect file")
	}
	file2.Close()

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
	defer f.Close()

	if f.Name() != name {
		t.Errorf("Wrong name. Want [%s], got: [%s]", name, f.Name())
	}
}

func TestFile_GetAndSetExpiresAt(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f, _ := Create(filepath.Join(tmpDir, "a"), 0664, nil)
	defer f.Close()

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

func TestFile_SetData(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f, _ := Create(filepath.Join(tmpDir, "a"), 0664, nil)
	defer f.Close()

	//content := []byte(strings.Repeat("Test", 2))
	content := []byte{111, 222}
	data := bytes.NewBuffer(content)

	exp := time.Now()
	if err := f.SetExpiresAt(exp); err != nil {
		t.Errorf("Got unexpected error on experation setting: %v", err)
	}

	if err := f.SetData(data); err != nil {
		t.Fatalf("Got unexpected error on data setting: %v", err)
	}

	_data, _ := ioutil.ReadAll(f.file)
	fmt.Println("==>", string(_data), _data)
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
