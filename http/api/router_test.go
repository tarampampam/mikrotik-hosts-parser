package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetRoutesHandlerFunc(t *testing.T) {
	var (
		req, _ = http.NewRequest("GET", "http://testing", nil)
		rr     = httptest.NewRecorder()
		router = mux.NewRouter()
	)

	router.
		HandleFunc("/gen", func(http.ResponseWriter, *http.Request) {}).
		Methods("GET").
		Name("script_generator")

	router.
		HandleFunc("/foo", func(http.ResponseWriter, *http.Request) {}).
		Methods("GET").
		Name("foo_bar") // must be skipped

	GetRoutesHandlerFunc(router)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response HTTP code. Want %d, got %d", http.StatusOK, rr.Code)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}

	if len(data) != 1 {
		t.Errorf("Wrong routes count. Expected length is 1, actual: %v", len(data))
	}

	if data["script_generator"].(map[string]interface{})["path"] != "/gen" {
		t.Errorf("Required route for `script_generator` was mot found in %v", data)
	}
}
