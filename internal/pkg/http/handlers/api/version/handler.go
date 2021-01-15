// Package version contains version API handler.
package version

import (
	"encoding/json"
	"net/http"
)

// NewHandler creates version handler.
func NewHandler(ver string) http.HandlerFunc {
	var cache []byte

	return func(w http.ResponseWriter, _ *http.Request) {
		if cache == nil {
			cache, _ = json.Marshal(struct {
				Version string `json:"version"`
			}{
				Version: ver,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(cache)
	}
}
