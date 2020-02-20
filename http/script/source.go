package script

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mikrotik-hosts-parser/settings/serve"
	"net/http"
	"time"
)

var httpClient = newHttpClient()

// RouterOsScriptSourceGenerationHandlerFunc generates RouterOS script source and writes it response.
func RouterOsScriptSourceGenerationHandlerFunc(serveSettings *serve.Settings) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParameters, queryErr := newQueryParametersBagUsingQueryValues(r.URL.Query())

		// Validate query parameters parsing
		if queryErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("## Query parameters error: " + queryErr.Error()))

			return
		}

		// stack with processing errors
		processingErrors := make([]string, 0)

		for _, sourceUrl := range queryParameters.sourceUrls {
			if response, err := fetchSourceContent(sourceUrl); err == nil {
				body, err := ioutil.ReadAll(io.LimitReader(response.Body, int64(serveSettings.RouterScript.MaxSourceSize)))

				if err != nil {
					processingErrors = append(processingErrors, sourceUrl + ": " + err.Error())

					continue
				}

				fmt.Println(body)
			} else {
				processingErrors = append(processingErrors, sourceUrl + ": " + err.Error())
			}
		}

		fmt.Println(processingErrors)
	}
}

func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: time.Second * 10, // Set request timeout
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 2 {
				return errors.New("request: too many (2) redirects")
			}
			return nil
		},
	}
}

func fetchSourceContent(uri string) (*http.Response, error) { // @todo: return `io.Reader` and use cache
	// Create HTTP request
	httpRequest, requestErr := http.NewRequest("GET", uri, nil)

	// Check request creation
	if requestErr != nil {
		return nil, requestErr
	}

	// Do request
	response, responseErr := httpClient.Do(httpRequest)

	// Check response getting
	if responseErr != nil {
		return nil, responseErr
	}

	return response, nil
}
