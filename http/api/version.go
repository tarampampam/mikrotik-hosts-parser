package api

import (
	"encoding/json"
	ver "mikrotik-hosts-parser/version"
	"net/http"
)

type (
	version struct {
		Version string `json:"version"`
	}
)

// GetVersion writes json response with version data into response writer.
func GetVersion(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(version{
		Version: ver.Version(),
	})
}
