package files

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestItem_GetAndSetWithoutHotBuffering(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	filePath := filepath.Join(tmpDir, "test-item")
	content := []byte(strings.Repeat("foo ", 512))
	data := bytes.NewBuffer(content)

	item := NewItem(filePath, "test-key", 0, 0)

	if err := item.Set(data); err != nil {
		t.Errorf("Got unexpected error on data SET: %v", err)
	}

	if !item.IsHit() {
		t.Error("Just created cache item should return true on `IsHit()` function calling")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := item.Get(buf); err != nil {
		t.Errorf("Got unexpected error on data GET: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf(
			"Got unexpected content from cache item. Want: %v (%s), got: %v (%s)",
			content,
			content,
			buf.Bytes(),
			buf.Bytes(),
		)
	}
}

func TestItem_GetAndSetWithHotBuffering(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)

	filePath := filepath.Join(tmpDir, "test-item")
	content := []byte(strings.Repeat("a", 512))
	data := bytes.NewBuffer(content)
	hotBufTTL := 50 * time.Millisecond

	item := NewItem(filePath, "foo", 512, hotBufTTL)

	if err := item.Set(data); err != nil {
		t.Errorf("Got unexpected error on data SET: %v", err)
	}

	// now we remove temp directory with file associated with cache item, and as we enable hot buffering - nothing
	// should be changed during TTL is not expired
	removeTempDir(t, tmpDir)

	if !item.IsHit() {
		t.Error("Just created cache item should return true on `IsHit()` function calling")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := item.Get(buf); err != nil {
		t.Errorf("Got unexpected error on data GET: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), content) {
		t.Errorf(
			"Got unexpected content from cache item. Want: %v (%s), got: %v (%s)",
			content,
			content,
			buf.Bytes(),
			buf.Bytes(),
		)
	}

	// but now lets wait for a TTL (double time - for reliability) expiring and then make same check again - result
	// must be different
	time.Sleep(hotBufTTL * 2)

	if item.IsHit() {
		t.Error("Expired cache item should return false on `IsHit()` function calling")
	}

	buf2 := bytes.NewBuffer([]byte{})
	wantErrorMessage := fmt.Sprintf("file [%s] cannot be opened", filePath)
	if err := item.Get(buf2); err == nil || err.Error() != wantErrorMessage {
		t.Errorf("Unexpected error on data getting. Want: %s, got: %v", wantErrorMessage, err)
	}

	if buf2.Len() != 0 {
		t.Errorf("Unexpected buffer length gor expired cache item: %d", buf2.Len())
	}
}

func TestItem_GetAndSetWithTooSmallHotBuffer(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)

	content := []byte(strings.Repeat("a", 512))
	data := bytes.NewBuffer(content)
	hotBufTTL := 50 * time.Millisecond

	item := NewItem(filepath.Join(tmpDir, "test-item"), "foo", 64, hotBufTTL) // hot buffer length must be less then content

	if err := item.Set(data); err != nil {
		t.Errorf("Got unexpected error on data SET: %v", err)
	}

	// now we remove temp directory with file associated with cache item, and as we enable hot buffering, but hot
	// buffer length is too small then content size - value from buffer should be skipped
	removeTempDir(t, tmpDir)

	if item.IsHit() {
		t.Error("`IsHit()` function calling must return false (hot buffer len is too small for stored content)")
	}

	// but now lets wait for a TTL (double time - for reliability) expiring and then make same check again - result
	// must be different
	time.Sleep(hotBufTTL * 2)

	if item.IsHit() {
		t.Error("Expired cache item should return false on `IsHit()` function calling")
	}

	buf := bytes.NewBuffer([]byte{})
	if err := item.Get(buf); err == nil {
		t.Error("Expected error is not returned")
	}

	if buf.Len() != 0 {
		t.Errorf("Unexpected buffer length gor expired cache item: %d", buf.Len())
	}
}

func TestItem_GetSetConcurrent(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	content := []byte(strings.Repeat("a", 8*1024))

	tests := []struct {
		name               string
		giveItem           *Item
		giveThreadsCount   int
		giveOperationCount int
		setup              func(t *testing.T, item *Item)
		assert             func(t *testing.T, item *Item)
	}{
		{
			name:               "Set with hot buffering",
			giveItem:           NewItem(filepath.Join(tmpDir, "a"), "a", 8*1024, 100*time.Microsecond),
			giveThreadsCount:   32,
			giveOperationCount: 16,
			assert: func(t *testing.T, item *Item) {
				if err := item.Set(bytes.NewBuffer(content)); err != nil {
					t.Errorf("Got unexpected error on data SET: %v", err)
				}
			},
		},
		{
			name:               "Set without hot buffering",
			giveItem:           NewItem(filepath.Join(tmpDir, "b"), "b", 0, 0),
			giveThreadsCount:   32,
			giveOperationCount: 16,
			assert: func(t *testing.T, item *Item) {
				if err := item.Set(bytes.NewBuffer(content)); err != nil {
					t.Errorf("Got unexpected error on data SET: %v", err)
				}
			},
		},
		{
			name:               "Get with hot buffering",
			giveItem:           NewItem(filepath.Join(tmpDir, "c"), "c", 8*1024, 100*time.Microsecond),
			giveThreadsCount:   32,
			giveOperationCount: 16,
			setup: func(t *testing.T, item *Item) {
				if err := item.Set(bytes.NewBuffer(content)); err != nil {
					t.Errorf("Got unexpected error on data SET: %v", err)
				}
			},
			assert: func(t *testing.T, item *Item) {
				buf := bytes.NewBuffer([]byte{})
				if err := item.Get(buf); err != nil {
					t.Errorf("Got unexpected error on data GET: %v", err)
				}

				if !bytes.Equal(buf.Bytes(), content) {
					t.Errorf("Got unexpected content from cache item. Want: %v, got: %v", content, buf.Bytes())
				}
			},
		},
		{
			name:               "Get without hot buffering",
			giveItem:           NewItem(filepath.Join(tmpDir, "d"), "d", 0, 0),
			giveThreadsCount:   32,
			giveOperationCount: 16,
			setup: func(t *testing.T, item *Item) {
				if err := item.Set(bytes.NewBuffer(content)); err != nil {
					t.Errorf("Got unexpected error on data SET: %v", err)
				}
			},
			assert: func(t *testing.T, item *Item) {
				buf := bytes.NewBuffer([]byte{})
				if err := item.Get(buf); err != nil {
					t.Errorf("Got unexpected error on data GET: %v", err)
				}

				if !bytes.Equal(buf.Bytes(), content) {
					t.Errorf("Got unexpected content from cache item. Want: %v, got: %v", content, buf.Bytes())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := sync.WaitGroup{}

			if tt.setup != nil {
				tt.setup(t, tt.giveItem)
			}

			for i := 0; i < tt.giveThreadsCount; i++ {
				wg.Add(1)
				go func() {
					for j := 0; j < tt.giveOperationCount; j++ {
						tt.assert(t, tt.giveItem)
					}
					wg.Done()
				}()
			}

			wg.Wait()
		})
	}
}

func TestItem_GetAndSetConcurrentWithoutHotBuffer(t *testing.T) { // nolint:gocyclo
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	tests := []struct {
		name     string
		giveItem *Item
		setup    func(t *testing.T, item *Item)
	}{
		{
			name:     "Get and set concurrent without hot buffer",
			giveItem: NewItem(filepath.Join(tmpDir, "a"), "a", 0, 0),
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
		{
			name:     "Get and set concurrent with hot buffer",
			giveItem: NewItem(filepath.Join(tmpDir, "b"), "b", 64, 2*time.Second),
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
		{
			name:     "Get and set concurrent with hot buffer with very low TTL",
			giveItem: NewItem(filepath.Join(tmpDir, "c"), "c", 64, 5*time.Nanosecond),
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
		{
			name:     "Get and set concurrent with hot buffer with small buffer",
			giveItem: NewItem(filepath.Join(tmpDir, "c"), "c", 6, 2*time.Second),
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

func TestItem_IsExpired(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)
	defer removeTempDir(t, tmpDir)

	item := NewItem(filepath.Join(tmpDir, "a"), "a", 0, 0)

	if ok, _ := item.IsExpired(); ok {
		t.Errorf("Just created item cannot be expirered")
	}

	expiresAt := time.Now().Add(10 * time.Millisecond)

	if err := item.SetExpiresAt(expiresAt); err != nil {
		t.Errorf("Unexpected error on expirind set: %v", err)
	}

	time.Sleep(11 * time.Millisecond)

	if ok, _ := item.IsExpired(); !ok {
		t.Error("Expired must return 'true' on `IsExpired` calling")
	}

	if item.ExpiresAt().UnixNano() != expiresAt.UnixNano() {
		t.Errorf("Wrong `ExpiredAt` result. Want %v, got: %v", expiresAt, item.ExpiresAt())
	}
}

func TestItem_ExpiringUsesHotBuffer(t *testing.T) {
	t.Parallel()

	tmpDir := createTempDir(t)

	hotBufferTTL := 10 * time.Millisecond
	item := NewItem(filepath.Join(tmpDir, "a"), "a", 0, hotBufferTTL)

	if ok, _ := item.IsExpired(); ok {
		t.Errorf("Just created item cannot be expirered")
	}

	expiresAt := time.Now().Add(hotBufferTTL)

	if err := item.SetExpiresAt(expiresAt); err != nil {
		t.Errorf("Unexpected error on expirind set: %v", err)
	}

	// After directory deleting data must be returned from hot cache
	removeTempDir(t, tmpDir)
	time.Sleep(hotBufferTTL / 2)

	if ok, _ := item.IsExpired(); ok {
		t.Error("Not expired must return 'false' on `IsExpired` calling")
	}

	if item.ExpiresAt().UnixNano() != expiresAt.UnixNano() {
		t.Errorf("Wrong `ExpiredAt` result. Want %v, got: %v", expiresAt, item.ExpiresAt())
	}

	// wait for hot cache expiring
	time.Sleep(hotBufferTTL)

	if _, err := item.IsExpired(); err == nil {
		t.Error("Expired item must return an 'error' on `IsExpired` calling")
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
