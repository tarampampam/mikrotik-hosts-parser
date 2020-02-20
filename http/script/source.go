package script

import (
	"encoding/json"
	"net/http"
)

// RouterOS script source generation handler.
func RouterOsScriptSourceGenerationHandler(w http.ResponseWriter, _ *http.Request) {
	res := make(map[string]interface{})
	// Append version
	res["work"] = "in progress"

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(res)
}
