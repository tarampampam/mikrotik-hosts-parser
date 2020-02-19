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

	routes map[string]route
)

// GetVersion writes json response with version data into response writer.
func GetRoutes(router *mux.Router, w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)

	response := routes{}

	for _, name := range routeNames {
		if path, err := router.Get(name).GetPathTemplate(); err == nil {
			response[name] = route{Path: path}
		}
	}

	_ = json.NewEncoder(w).Encode(response)
}
