package files

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestItem_GetAndSetWithoutHotBuffering(t *testing.T) {
	t.Parallel()

	// Create temporary file inside just created temporary directory.
	tmpDir, err := ioutil.TempDir("", "test-")
	if err != nil {
		t.Fatal(err)
	}

	// Remove temporary directory when test is completed
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	}()

	filePath := filepath.Join(tmpDir, "test-item")
	content := []byte("some content")
	data := bytes.NewBuffer(content)
	item := NewItem(filePath, "foo", 0, 0)

	if err = item.Set(data); err != nil {
		t.Errorf("Got unexpected error on data SET: %v", err)
	}

	if !item.IsHit() {
		t.Error("Just created cache item should return true on `IsHit()` function calling")
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	if err := item.Get(buf); err != nil {
		t.Errorf("Got unexpected error on data GET: %v", err)
	}

	if bytes.Compare(buf.Bytes(), content) != 0 {
		t.Errorf(
			"Got unexpected content from cache item. Want: %v (%s), got: %v (%s)",
			content,
			content,
			buf.Bytes(),
			buf.Bytes(),
		)
	}
}
