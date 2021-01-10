package version

import (
	"encoding/json"
	"net/http"
)

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

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(cache)
	}
}