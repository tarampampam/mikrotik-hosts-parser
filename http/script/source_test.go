package script

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterOsScriptSourceGenerationHandler(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
	)

	RouterOsScriptSourceGenerationHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response HTTP code. Want %d, got %d", http.StatusOK, rr.Code)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
}
