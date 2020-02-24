package script

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type httpClient struct {
	client *http.Client
}

var defaultHTTPClient = newHTTPClient()

// newHTTPClient creates new http client for a working with external sources
func newHTTPClient() *httpClient {
	return &httpClient{
		client: &http.Client{
			Timeout: time.Second * 10, // Set request timeout
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= 2 {
					return errors.New("request: too many (2) redirects")
				}

				return nil
			},
		},
	}
}

// FetchSourceContent sends request to the external source and returns response object only if ALL is ok
func (c *httpClient) FetchSourceContent(uri string, maxLength int) (*http.Response, error) {
	// Create HTTP request
	httpRequest, requestErr := http.NewRequest("GET", uri, nil)

	// Check request creation
	if requestErr != nil {
		return nil, requestErr
	}

	// Do request
	response, responseErr := c.client.Do(httpRequest)

	// Check response getting
	if responseErr != nil {
		return nil, responseErr
	}

	// `Content-Type` header validation
	if contentType := response.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		_ = response.Body.Close()

		return nil, fmt.Errorf("wrong 'Content-Type' header '%s', required is '%s'", contentType, "text/plain*")
	}

	// `Content-Length` header validation (if last presents)
	if contentLength := response.Header.Get("Content-Length"); contentLength != "" {
		value, parseErr := strconv.Atoi(contentLength)

		// Parse value
		if parseErr != nil {
			_ = response.Body.Close()

			return nil, errors.New("header 'Content-Length' parsing error: " + parseErr.Error())
		}

		// Validate length
		if value >= maxLength {
			_ = response.Body.Close()

			return nil, fmt.Errorf("'Content-Length' header value is too much (%d, maximum: %d)", value, maxLength)
		}
	}

	// @todo: add read response body size checking

	return response, nil
}
