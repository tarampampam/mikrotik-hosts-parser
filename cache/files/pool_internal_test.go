package files

import (
	"testing"
	"time"
)

func TestPool_keyToHash(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	pool := NewPool(tmpDir, 128, time.Second)

	tests := []struct {
		giveKey  string
		wantHash string
	}{
		{giveKey: "foo", wantHash: "acbd18db4cc2f85cedef654fccc4a4d8"},
		{giveKey: "bar", wantHash: "37b51d194a7513e45b56f6524f2d51f2"},
		{giveKey: "@$^#J&LKNRB(CBSsd0hs)_$iJ4n^^YFSH03", wantHash: "587e8aa537ee439dced3ee2ba01e6940"},
	}

	var lastHash string

	for _, tt := range tests {
		hash := pool.keyToHash(tt.giveKey)

		if hash != tt.wantHash {
			t.Errorf("Wrong key hashing for gived key (%s). Want: %v, got: %v", tt.giveKey, tt.wantHash, hash)
		}

		if lastHash == hash {
			t.Error("Identical hash with previous hashing function call detected")
		}

		lastHash = hash
	}
}
