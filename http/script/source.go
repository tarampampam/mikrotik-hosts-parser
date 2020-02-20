package script

import (
	"encoding/json"
	"errors"
	"fmt"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// query parameters bag structure
type queryParametersBag struct {
	sourceUrls    []string
	format        string
	version       string
	excludedHosts []string
	limit         int
	redirectTo    string
}

type sourceFetcher struct{}

// RouterOsScriptSourceGenerationHandlerFunc generates RouterOS script source and writes it response.
func RouterOsScriptSourceGenerationHandlerFunc(serveSettings *serve.Settings) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var queryParameters = queryParametersBag{}

		// Parse query parameters into query parameters bag
		if err := queryParameters.initUsingQueryValues(r.URL.Query()); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("## Query parameters error: " + err.Error()))

			return
		}

		// Make sure that sources list is not empty
		if len(queryParameters.sourceUrls) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("## Error: empty sources list"))

			return
		}

		fmt.Println(queryParameters)

		res := make(map[string]interface{})
		// Append version
		res["work"] = "in progress"

		w.WriteHeader(http.StatusOK)

		_ = json.NewEncoder(w).Encode(res)
	}
}

//func fetchSourceContent(uri string) ([]byte, error) {
//
//}

func (bag *queryParametersBag) initUsingQueryValues(values url.Values) error {
	// Extract `sources_urls` values
	if sourceUrls, ok := values["sources_urls"]; ok {
		// Iterate query values slice
		for _, value := range sourceUrls {
			// Explode value with URLs list (separated using `,`) into single URLs
			for _, sourceUrl := range strings.Split(value, ",") {
				// Make URL validation, and if all is ok - append it into query parameters bag
				if _, err := url.ParseRequestURI(sourceUrl); err == nil {
					bag.sourceUrls = append(bag.sourceUrls, sourceUrl)
				}
			}
		}
	} else {
		return errors.New("required parameter `sources_urls` was not found")
	}

	// Extract `format` value
	if value, ok := values["format"]; ok {
		if len(value) > 0 {
			bag.format = value[0]
		}
	}

	// Extract `format` value
	if value, ok := values["version"]; ok {
		if len(value) > 0 {
			bag.version = value[0]
		}
	}

	// Extract `excluded_hosts` value
	if excludedHosts, ok := values["excluded_hosts"]; ok {
		// Iterate query values slice
		for _, value := range excludedHosts {
			// Explode value with host names list (separated using `,`) into single host names
			for _, excludedHost := range strings.Split(value, ",") {
				// Make basic checking, and if all is ok - append it into query parameters bag
				if excludedHost != "" {
					bag.excludedHosts = append(bag.excludedHosts, excludedHost)
				}
			}
		}
	}

	// Extract `limit` value
	if value, ok := values["limit"]; ok {
		if len(value) > 0 {
			if value, err := strconv.Atoi(value[0]); err == nil {
				bag.limit = value
			} else {
				return errors.New("wrong `limit` value (cannot be converted into integer)")
			}
		}
	}

	// Extract `redirect_to` value
	if value, ok := values["redirect_to"]; ok {
		if len(value) > 0 {
			bag.redirectTo = value[0]
		}
	}

	return nil
}
