package files

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestItem_GetAndSet(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	key := "test-key"
	content := []byte(strings.Repeat("foo ", 32))
	item := NewItem(tmpDir, key)

	if item.IsHit() {
		t.Error("Just created cache item should return false on `IsHit()` function calling")
	}

	if err := item.Set(bytes.NewBuffer(content)); err != nil {
		t.Errorf("Got unexpected error on data SET: %v", err)
	}

	if !item.IsHit() {
		t.Error("Created cache item should return true on `IsHit()` function calling")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := item.Get(buf); err != nil {
		t.Errorf("Got unexpected error on data GET: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf("Got unexpected content from cache item. Want: %v, got: %v", content, buf.Bytes())
	}
}

func TestItem_GetAndSetConcurrent(t *testing.T) { // nolint:gocyclo
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	tests := []struct {
		name     string
		giveItem *Item
		setup    func(t *testing.T, item *Item)
	}{
		{
			name:     "Default set and set concurrent",
			giveItem: NewItem(tmpDir, "a"),
			setup: func(t *testing.T, item *Item) {
				// setup basic state
				if err := item.Set(bytes.NewBuffer([]byte(strings.Repeat("x", 32)))); err != nil {
					t.Errorf("Got unexpected error on data SET: %v", err)
				}
				if err := item.SetExpiresAt(time.Now().Add(time.Second * 10)); err != nil {
					t.Errorf("Got unexpected error on set expiring time: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := sync.WaitGroup{}

			tt.setup(t, tt.giveItem)

			// start "Get" go routines
			for i := 0; i < 256; i++ {
				wg.Add(1)
				go func(item *Item) {
					defer wg.Done()
					if err := item.Get(bytes.NewBuffer([]byte{})); err != nil {
						t.Errorf("Got unexpected error on data GET: %v", err)
					}
					if !item.IsHit() {
						t.Error("Cache item should return true on `IsHit()` function calling")
					}
					if key := item.GetKey(); key != tt.giveItem.key {
						t.Errorf("Wrong key returged. Want: %s, got: %s", tt.giveItem.key, key)
					}
					if expTime := item.ExpiresAt(); expTime == nil {
						t.Error("Expiration time was not returned")
					}
					if _, err := item.IsExpired(); err != nil {
						t.Errorf("Got unexpected error on expiring checking: %v", err)
					}
				}(tt.giveItem)
			}

			// start "Set" go routines
			for i := 0; i < 256; i++ {
				wg.Add(1)
				go func(item *Item) {
					defer wg.Done()
					if err := item.Set(bytes.NewBuffer([]byte(strings.Repeat("z", 32)))); err != nil {
						t.Errorf("Got unexpected error on data SET: %v", err)
					}
					if !item.IsHit() {
						t.Error("Cache item should return true on `IsHit()` function calling")
					}
					if key := item.GetKey(); key != tt.giveItem.key {
						t.Errorf("Wrong key returged. Want: %s, got: %s", tt.giveItem.key, key)
					}
					if err := item.SetExpiresAt(time.Now().Add(time.Second * 10)); err != nil {
						t.Errorf("Got unexpected error on set expiring time: %v", err)
					}
				}(tt.giveItem)
			}

			wg.Wait()
		})
	}
}

func TestItem_ExpiringGetSetAndCheck(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	item := NewItem(tmpDir, "a")

	if ok, isExpErr := item.IsExpired(); ok {
		t.Errorf("Just created item cannot be expirered")
	} else if isExpErr == nil {
		t.Errorf("Expected error on expirind checking was not returned")
	}

	expiresAt := time.Now().Add(3 * time.Millisecond)

	if err := item.SetExpiresAt(expiresAt); err != nil {
		t.Errorf("Unexpected error on expirind set: %v", err)
	}

	// wait for expiring
	time.Sleep(4 * time.Millisecond)

	if ok, isExpErr := item.IsExpired(); !ok {
		t.Error("Expired must return 'true' on `IsExpired` calling")
	} else if isExpErr != nil {
		t.Errorf("Unexpected error on expirind checking: %v", isExpErr)
	}

	if item.ExpiresAt().Unix() != expiresAt.Unix() {
		t.Errorf("Wrong `ExpiredAt` result. Want %v, got: %v", expiresAt, item.ExpiresAt())
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
