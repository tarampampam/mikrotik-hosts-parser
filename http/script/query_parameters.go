package script

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

type queryParametersBag struct {
	SourceUrls    []string
	Format        string
	Version       string
	ExcludedHosts []string
	Limit         int
	RedirectTo    string
}

// newQueryParametersBag makes query parameters bag using passed url values.
func newQueryParametersBag( //nolint:gocyclo
	values url.Values,
	defaultRedirectTo string,
	maxSources int,
) (*queryParametersBag, error) {
	bag := &queryParametersBag{
		// Defaults:
		RedirectTo: defaultRedirectTo,
		Format:     "routeros",
	}

	// Extract `sources_urls` values
	if sourceUrls, ok := values["sources_urls"]; ok {
		// Iterate query values slice
		for _, value := range sourceUrls {
			// Explode value with URLs list (separated using `,`) into single URLs
			for _, sourceURL := range strings.Split(value, ",") {
				// Make URL validation, and if all is ok - append it into query parameters bag
				if _, err := url.ParseRequestURI(sourceURL); err == nil && len(sourceURL) <= 256 {
					bag.SourceUrls = append(bag.SourceUrls, sourceURL)
				}
			}
		}
	} else {
		return nil, errors.New("required parameter `sources_urls` was not found")
	}

	// Validate sources list size
	if len(bag.SourceUrls) < 1 {
		return nil, errors.New("empty sources list")
	}

	// remove duplicated sources
	bag.SourceUrls = bag.uniqueStringsSlice(bag.SourceUrls)

	// check for sources count
	if len(bag.SourceUrls) > maxSources {
		return nil, fmt.Errorf("too much sources (only %d is allowed)", maxSources)
	}

	// Extract `format` value
	if value, ok := values["format"]; ok {
		if len(value) > 0 {
			bag.Format = value[0]
		}
	}

	// Extract `version` value
	if value, ok := values["version"]; ok {
		if len(value) > 0 {
			bag.Version = value[0]
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
					bag.ExcludedHosts = append(bag.ExcludedHosts, excludedHost)
				}
			}
		}
		// remove duplicated hosts
		bag.ExcludedHosts = bag.uniqueStringsSlice(bag.ExcludedHosts)
		// Validate excluded hosts list size
		if len(bag.ExcludedHosts) > 32 {
			return nil, errors.New("too many excluded hosts (more then 32)")
		}
	}

	// Extract `limit` value
	if value, ok := values["limit"]; ok {
		if len(value) > 0 {
			if value, err := strconv.Atoi(value[0]); err == nil {
				if value <= 0 {
					return nil, errors.New("wrong `limit` value (cannot be less then 1)")
				}
				bag.Limit = value
			} else {
				return nil, errors.New("wrong `limit` value (cannot be converted into integer)")
			}
		}
	}

	// Extract and validate `redirect_to` value
	if value, ok := values["redirect_to"]; ok {
		if len(value) > 0 {
			if net.ParseIP(value[0]) == nil {
				return nil, errors.New("wrong `redirect_to` value (invalid IP address)")
			}
			bag.RedirectTo = value[0]
		}
	}

	return bag, nil
}

// uniqueStringsSlice removes duplicated strings from strings slice
func (queryParametersBag) uniqueStringsSlice(in []string) []string {
	keys := make(map[string]bool)
	out := make([]string, 0)

	for _, entry := range in {
		if _, ok := keys[entry]; !ok {
			keys[entry] = true
			out = append(out, entry)
		}
	}

	return out
}
