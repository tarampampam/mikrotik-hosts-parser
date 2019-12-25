package files

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

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
