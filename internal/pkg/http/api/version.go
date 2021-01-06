package api

import (
	"encoding/json"
	"net/http"

	ver "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/version"
)

type (
	version struct {
		Version string `json:"version"`
	}
)

// GetVersionHandler writes json response with version data into response writer.
func GetVersionHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(version{
		Version: ver.Version(),
	})
}
