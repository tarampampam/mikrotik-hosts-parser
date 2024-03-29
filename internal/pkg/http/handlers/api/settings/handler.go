// Package settings contains API handler for application settings getting.
package settings

import (
	"encoding/json"
	"net/http"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/cache"
	"gh.tarampamp.am/mikrotik-hosts-parser/v4/internal/pkg/config"
)

type (
	response struct {
		Sources struct {
			Provided      []providedSource `json:"provided"`
			Max           int              `json:"max"`
			MaxSourceSize int              `json:"max_source_size"` // in bytes
		} `json:"sources"`
		Redirect struct {
			Addr string `json:"addr"`
		} `json:"redirect"`
		Records struct {
			Comment string `json:"comment"`
		} `json:"records"`
		Excludes struct {
			Hosts []string `json:"hosts"`
		} `json:"excludes"`
		Cache struct {
			LifetimeSec int `json:"lifetime_sec"`
		} `json:"cache"`
	}

	providedSource struct {
		URI         string `json:"uri"`
		Name        string `json:"name"`
		Description string `json:"description"`
		ByDefault   bool   `json:"default"`
		Count       int    `json:"count"`
	}
)

// NewHandler creates settings handler.
func NewHandler(cfg config.Config, cacher cache.Cacher) http.HandlerFunc {
	var c []byte // response in-memory cache

	return func(w http.ResponseWriter, _ *http.Request) {
		if c == nil {
			// set basic properties
			resp := &response{}
			resp.Sources.Max = int(cfg.RouterScript.MaxSourcesCount)
			resp.Sources.MaxSourceSize = int(cfg.RouterScript.MaxSourceSizeBytes)
			resp.Redirect.Addr = cfg.RouterScript.Redirect.Address
			resp.Records.Comment = cfg.RouterScript.Comment
			resp.Cache.LifetimeSec = int(cacher.TTL().Seconds())

			// append excluded hosts list
			resp.Excludes.Hosts = append(resp.Excludes.Hosts, cfg.RouterScript.Exclude.Hosts...)

			// append sources list entries
			for _, source := range cfg.Sources {
				resp.Sources.Provided = append(resp.Sources.Provided, providedSource{
					URI:         source.URI,
					Name:        source.Name,
					Description: source.Description,
					ByDefault:   source.EnabledByDefault,
					Count:       int(source.RecordsCount),
				})
			}

			c, _ = json.Marshal(resp)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(c)
	}
}
