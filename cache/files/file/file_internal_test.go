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

func TestFTypeOffsetConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		constName string
		giveConst fOffset
		wantValue uint16
	}{
		{constName: "oFTypeFrom", giveConst: oFTypeFrom, wantValue: 0},
		{constName: "oFTypeTo", giveConst: oFTypeTo, wantValue: 7},
		{constName: "oFMetaFrom", giveConst: oFMetaFrom, wantValue: 8},
		{constName: "oFMetaExpAtUnixMsFrom", giveConst: oFMetaExpAtUnixMsFrom, wantValue: 8},
		{constName: "oFMetaExpAtUnixMsTo", giveConst: oFMetaExpAtUnixMsTo, wantValue: 22},
		{constName: "oFMetaTo", giveConst: oFMetaTo, wantValue: 247},
		{constName: "oFDataSHA1From", giveConst: oFDataSHA1From, wantValue: 248},
		{constName: "oFDataSHA1To", giveConst: oFDataSHA1To, wantValue: 288},
		{constName: "oFDataFrom", giveConst: oFDataFrom, wantValue: 289},
	}

	for _, tt := range tests {
		t.Run(tt.constName, func(t *testing.T) {
			if uint16(tt.giveConst) != tt.wantValue {
				t.Errorf("Wrong value for constant '%s'. Want: %v, got: %v", tt.constName, tt.wantValue, tt.wantValue)
			}
		})
	}
}

func Test_getBlockPosition(t *testing.T) {
	t.Parallel()

	const unknown blockType = 255

	tests := []struct {
		blockName     string
		giveBlockType blockType
		wantFrom      fOffset
		wantTo        fOffset
	}{
		{blockName: "bFType", giveBlockType: bFType, wantFrom: oFTypeFrom, wantTo: oFTypeTo},
		{blockName: "bFMeta", giveBlockType: bFMeta, wantFrom: oFMetaFrom, wantTo: oFMetaTo},
		{blockName: "bFMetaExpAtUnixMS", giveBlockType: bFMetaExpAtUnixMS, wantFrom: oFMetaExpAtUnixMsFrom, wantTo: oFMetaExpAtUnixMsTo},
		{blockName: "bFDataSHA1", giveBlockType: bFDataSHA1, wantFrom: oFDataSHA1From, wantTo: oFDataSHA1To},
		{blockName: "unknown", giveBlockType: unknown, wantFrom: 0, wantTo: 0},
	}

	for _, tt := range tests {
		t.Run(tt.blockName, func(t *testing.T) {
			from, to := (&File{}).getBlockPosition(tt.giveBlockType)
			if from != tt.wantFrom {
				t.Errorf("Wrong 'from' for type '%s'. Want: %v, got: %v", tt.blockName, tt.wantFrom, from)
			}
			if to != tt.wantTo {
				t.Errorf("Wrong 'to' for type '%s'. Want: %v, got: %v", tt.blockName, tt.wantTo, to)
			}
		})
	}
}

func TestFile_GetAndSetType(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	const fakeType FType = "X1234567"

	tests := []struct {
		giveType FType
		wantType FType
	}{
		{giveType: TUnknown, wantType: TUnknown},
		{giveType: TRegularCacheEntry, wantType: TRegularCacheEntry},
		{giveType: fakeType, wantType: TUnknown},
	}

	for _, tt := range tests {
		t.Run("With "+string(tt.giveType), func(t *testing.T) {
			f, createErr := Create(filepath.Join(tmpDir, string(tt.giveType)), 0664)

			if createErr != nil {
				t.Fatalf("Got unexpected error on file creation: %v", createErr)
			}

			fType, getErr := f.GetType()
			if getErr != nil {
				t.Errorf("Got unexpected error on type getting: %v", getErr)
			}

			if fType != TUnknown {
				t.Errorf("Unexpected type returned. Want: %v, got: %v", TUnknown, fType)
			}

			if setErr := f.SetType(tt.giveType); setErr != nil {
				t.Errorf("Got unexpected error on type setting: %v", setErr)
			}

			fType, getErr = f.GetType()
			if getErr != nil {
				t.Errorf("Got unexpected error on type getting: %v", getErr)
			}

			if tt.wantType != fType {
				data, _ := ioutil.ReadAll(f.file)
				t.Errorf("Unexpected type returned. Want: %v, got: %v. File content: %v (%s)", tt.wantType, fType, data, data)
			}

			if closeErr := f.Close(); closeErr != nil {
				t.Fatalf("Got unexpected error on file closing: %v", closeErr)
			}
		})
	}
}

func TestFile_SetTypeWithWrongValue(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	const wrongType FType = "X12345678"

	f, createErr := Create(filepath.Join(tmpDir, string(wrongType)), 0664)
	if createErr != nil {
		t.Fatalf("Got unexpected error on file creation: %v", createErr)
	}
	defer f.Close()

	if setErr := f.SetType(wrongType); setErr == nil {
		t.Error("Expected error was not returned")
	}

	fType, _ := f.GetType()
	if fType != TUnknown {
		t.Errorf("Got unexpected error on type getting: %v. Want: %v", fType, TUnknown)
	}
}

func TestFile_GetAndSetExpiresAt(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f, createErr := Create(filepath.Join(tmpDir, "a"), 0664)
	if createErr != nil {
		t.Fatalf("Got unexpected error on file creation: %v", createErr)
	}
	defer f.Close()

	if v, err := f.GetExpiresAt(); err == nil {
		t.Errorf("Expected error does not returned (value is %v)", v)
	}

	if tp, _ := f.GetType(); tp != TUnknown {
		t.Errorf("Wrong type is set by default: %v", tp)
	}

	exp := time.Now()

	if err := f.SetExpiresAt(exp); err != nil {
		t.Fatalf("Got unexpected error on experation setting: %v", err)
	}

	gotExp, getErr := f.GetExpiresAt()
	if getErr != nil {
		t.Fatalf("Got unexpected error on experation getting: %v", getErr)
	}

	if gotExp.Unix() != exp.Unix() {
		t.Errorf("Got wrong experation time value. Want: %v, got: %v", exp, gotExp)
	}
}

func TestFile_hashSha1Methods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		giveString string
		wantError bool
	}{
		{giveString: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde", wantError: false},
		{giveString: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcd€", wantError: true}, // `€` is [226 130 172] - 3 bytes
		{giveString: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde11", wantError: true}, // too long
		{giveString: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabc", wantError: true}, // too short
		{giveString: "", wantError: true}, // too short
	}

	for _, tt := range tests {
		t.Run(tt.giveString, func(t *testing.T) {
			h, err := newHashSha1([]byte(tt.giveString))

			if tt.wantError && err == nil {
				t.Error("Expects error, but nothing returned")
			}

			if !tt.wantError && h.String() != tt.giveString {
				t.Errorf("Wrong hash to string convertation. Want %v, got %v", tt.giveString, h)
			}
		})
	}
}

func TestFile_SetData(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	f, createErr := Create(filepath.Join(tmpDir, "a"), 0664)
	if createErr != nil {
		t.Fatalf("Got unexpected error on file creation: %v", createErr)
	}
	defer f.Close()

	content := []byte(strings.Repeat("Test", 2))
	data := bytes.NewBuffer(content)

	if err := f.SetData(data); err != nil {
		t.Fatalf("Got unexpected error on data setting: %v", err)
	}

	_data, _ := ioutil.ReadAll(f.file)
	fmt.Println(_data)
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
