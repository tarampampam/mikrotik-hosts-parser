package script

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type queryParametersBag struct {
	sourceUrls    []string
	format        string
	version       string
	excludedHosts []string
	limit         int
	redirectTo    string
}

// newQueryParametersBagUsingQueryValues makes query parameters bag using passed url values.
func newQueryParametersBagUsingQueryValues(values url.Values) (*queryParametersBag, error) {
	bag := &queryParametersBag{}

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
		return nil, errors.New("required parameter `sources_urls` was not found")
	}

	// Validate sources list size
	if len(bag.sourceUrls) < 1 {
		return nil, errors.New("empty sources list")
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
				return nil, errors.New("wrong `limit` value (cannot be converted into integer)")
			}
		}
	}

	// Extract `redirect_to` value
	if value, ok := values["redirect_to"]; ok {
		if len(value) > 0 {
			bag.redirectTo = value[0]
		}
	}

	return bag, nil
}
