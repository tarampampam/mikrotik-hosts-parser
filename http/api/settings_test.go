package api

import (
	"encoding/json"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSettingsHandlerFunc(t *testing.T) { //nolint:gocyclo
	var (
		req, _        = http.NewRequest("GET", "http://testing", nil)
		rr            = httptest.NewRecorder()
		serveSettings = serve.Settings{
			Sources: []serve.Source{{
				URI:              "http://goo.gl/hosts.txt",
				Name:             "Foo name",
				Description:      "Foo desc",
				EnabledByDefault: true,
				RecordsCount:     123,
			}, {
				URI:              "http://face.book/hosts.txt",
				Name:             "Bar name",
				Description:      "Bar desc",
				EnabledByDefault: false,
				RecordsCount:     -321,
			}},
			RouterScript: serve.RouterScript{
				Redirect: serve.Redirect{
					Address: "0.1.1.0",
				},
				Exclude: serve.Excludes{
					Hosts: []string{"foo", "bar"},
				},
				MaxSources:    1,
				Comment:       " [ blah ] ",
				MaxSourceSize: 666,
			},
			Cache: serve.Cache{
				LifetimeSec: 1234,
			},
		}
	)

	GetSettingsHandlerFunc(&serveSettings)(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Wrong response HTTP code. Want %d, got %d", http.StatusOK, rr.Code)
	}

	data := make(map[string]interface{})
	if err := json.Unmarshal(rr.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}

	var (
		sourcesProvided  = data["sources"].(map[string]interface{})["provided"].([]interface{})
		sourcesMax       = int(data["sources"].(map[string]interface{})["max"].(float64))
		maxSourceSize    = int(data["sources"].(map[string]interface{})["max_source_size"].(float64))
		redirectAddr     = data["redirect"].(map[string]interface{})["addr"].(string)
		recordsComment   = data["records"].(map[string]interface{})["comment"].(string)
		cacheLifetimeSec = int(data["cache"].(map[string]interface{})["lifetime_sec"].(float64))
		excludesHosts    = data["excludes"].(map[string]interface{})["hosts"].([]interface{})
	)

	if len(serveSettings.Sources) != len(sourcesProvided) {
		t.Errorf("Wrong source records count. Want: %v, got: %v", len(serveSettings.Sources), len(sourcesProvided))
	}

	for i := range serveSettings.Sources {
		if sourcesProvided[i].(map[string]interface{})["name"] != serveSettings.Sources[i].Name {
			t.Errorf("Unexpected source name found in: %v", sourcesProvided[i])
		}
		if int(sourcesProvided[i].(map[string]interface{})["count"].(float64)) != serveSettings.Sources[i].RecordsCount {
			t.Errorf("Unexpected records count found in: %v", sourcesProvided[i])
		}
		if sourcesProvided[i].(map[string]interface{})["default"] != serveSettings.Sources[i].EnabledByDefault {
			t.Errorf("Unexpected default value found in: %v", sourcesProvided[i])
		}
		if sourcesProvided[i].(map[string]interface{})["description"] != serveSettings.Sources[i].Description {
			t.Errorf("Unexpected source description found in: %v", sourcesProvided[i])
		}
		if sourcesProvided[i].(map[string]interface{})["uri"] != serveSettings.Sources[i].URI {
			t.Errorf("Unexpected URI found in: %v", sourcesProvided[i])
		}
	}

	if sourcesMax != serveSettings.RouterScript.MaxSources {
		t.Errorf("Unexpected max sources: got %v, want %v", sourcesMax, serveSettings.RouterScript.MaxSources)
	}

	if maxSourceSize != serveSettings.RouterScript.MaxSourceSize {
		t.Errorf("Unexpected max source size: got %v, want %v", maxSourceSize, serveSettings.RouterScript.MaxSourceSize)
	}

	if redirectAddr != serveSettings.RouterScript.Redirect.Address {
		t.Errorf("Unexpected redirect address comment: got %v, want %v", redirectAddr, serveSettings.RouterScript.Redirect.Address)
	}

	if cacheLifetimeSec != serveSettings.Cache.LifetimeSec {
		t.Errorf("Unexpected cache lifetime: got %v, want %v", cacheLifetimeSec, serveSettings.Cache.LifetimeSec)
	}

	if recordsComment != serveSettings.RouterScript.Comment {
		t.Errorf("Unexpected records comment: got %v, want %v", recordsComment, serveSettings.RouterScript.Comment)
	}

	for i := range serveSettings.RouterScript.Exclude.Hosts {
		if excludesHosts[i] != serveSettings.RouterScript.Exclude.Hosts[i] {
			t.Errorf("Unexpected excluded host found: %v", excludesHosts[i])
		}
	}
}
