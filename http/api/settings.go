package api

import (
	"encoding/json"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
)

type (
	providedSource struct {
		URI         string `json:"uri"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ByDefault   bool   `json:"default"`
		Count       int    `json:"count"`
	}

	sources struct {
		Provided      []providedSource `json:"provided"`
		Max           int              `json:"max"`
		MaxSourceSize int              `json:"max_source_size"` // in bytes
	}

	redirect struct {
		Addr string `json:"addr"`
	}

	records struct {
		Comment string `json:"comment"`
	}

	excludes struct {
		Hosts []string `json:"hosts"`
	}

	settingsResponse struct {
		Sources  sources  `json:"sources"`
		Redirect redirect `json:"redirect"`
		Records  records  `json:"records"`
		Excludes excludes `json:"excludes"`
		Cache    cache    `json:"cache"`
	}

	cache struct {
		LifetimeSec int `json:"lifetime_sec"`
	}
)

// GetSettingsHandlerFunc returns handler function that writes json response with possible settings into response writer.
func GetSettingsHandlerFunc(serveSettings *serve.Settings) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(convertServeSettingsIntoResponse(serveSettings))
	}
}

// convertServeSettingsIntoResponse converts serving settings into internal response format.
func convertServeSettingsIntoResponse(settings *serve.Settings) *settingsResponse {
	// set basic properties
	response := &settingsResponse{
		Sources: sources{
			Max:           settings.RouterScript.MaxSources,
			MaxSourceSize: settings.RouterScript.MaxSourceSize,
		},
		Redirect: redirect{
			Addr: settings.RouterScript.Redirect.Address,
		},
		Records: records{
			Comment: settings.RouterScript.Comment,
		},
		Cache: cache{
			LifetimeSec: settings.Cache.LifetimeSec,
		},
		Excludes: excludes{},
	}

	// append excluded hosts list
	response.Excludes.Hosts = append(response.Excludes.Hosts, settings.RouterScript.Exclude.Hosts...)

	// append sources list entries
	for _, source := range settings.Sources {
		response.Sources.Provided = append(response.Sources.Provided, providedSource{
			URI:         source.URI,
			Name:        source.Name,
			Description: source.Description,
			ByDefault:   source.EnabledByDefault,
			Count:       source.RecordsCount,
		})
	}

	return response
}
