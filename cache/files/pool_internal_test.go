package files

import "testing"

func TestNewPool(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir)

	if pool.cacheDirPath != tmpDir {
		t.Errorf("Unexpected cache dir path: %v", tmpDir)
	}
}
