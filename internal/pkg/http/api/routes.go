package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// supported route names
var routeNames = [...]string{"script_generator"}

type (
	route struct {
		Path string `json:"path"`
	}
)

// GetSettingsHandlerFunc returns handler function that writes json response with version data into response writer.
func GetRoutesHandlerFunc(router *mux.Router) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		response := make(map[string]route)

		for _, name := range routeNames {
			if path, err := router.Get(name).GetPathTemplate(); err == nil {
				response[name] = route{Path: path}
			}
		}

		_ = json.NewEncoder(w).Encode(response)
	}
}
