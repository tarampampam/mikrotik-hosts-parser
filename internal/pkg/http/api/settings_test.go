package api

/*
func TestGetSettingsHandlerFunc(t *testing.T) { //nolint:gocyclo
	var (
		req, _        = http.NewRequest("GET", "http://testing", nil)
		rr            = httptest.NewRecorder()
		serveSettings = config.ServingConfig{
			Sources: []config.source{{
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
			RouterScript: config.routerScript{
				Redirect: config.redirect{
					Address: "0.1.1.0",
				},
				Exclude: config.excludes{
					Hosts: []string{"foo", "bar"},
				},
				MaxSourcesCount:    1,
				Comment:            " [ blah ] ",
				MaxSourceSizeBytes: 666,
			},
			Cache: config.cache{
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

	if sourcesMax != serveSettings.RouterScript.MaxSourcesCount {
		t.Errorf("Unexpected max sources: got %v, want %v", sourcesMax, serveSettings.RouterScript.MaxSourcesCount)
	}

	if maxSourceSize != serveSettings.RouterScript.MaxSourceSizeBytes {
		t.Errorf("Unexpected max source size: got %v, want %v", maxSourceSize, serveSettings.RouterScript.MaxSourceSizeBytes)
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
*/
