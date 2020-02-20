package api

import (
	"encoding/json"
	ver "mikrotik-hosts-parser/version"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetVersionHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	GetVersionHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response HTTP code. Want %d, got %d", http.StatusOK, rr.Code)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}

	if version, _ := data["version"].(string); version != ver.Version() {
		t.Errorf("unexpected version: got %v want %v", version, ver.Version())
	}
}
