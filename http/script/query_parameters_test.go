package script

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Test_newQueryParametersBag(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		name                   string
		giveURLValues          url.Values
		giveDefaultRedirectTo  string
		giveMaxSources         int
		wantError              error
		wantQueryParametersBag queryParametersBag
	}{
		{
			name: "Basic usage",
			giveURLValues: url.Values{
				"format":         []string{"routeros"},
				"version":        []string{"foo@bar"},
				"redirect_to":    []string{"127.0.0.10"},
				"limit":          []string{"50"},
				"sources_urls":   []string{"http://foo.com/bar.txt,http://bar.com/baz.asp"},
				"excluded_hosts": []string{"foo.com,bar.com"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls:    []string{"http://foo.com/bar.txt", "http://bar.com/baz.asp"},
				Format:        "routeros",
				Version:       "foo@bar",
				ExcludedHosts: []string{"foo.com", "bar.com"},
				Limit:         50,
				RedirectTo:    "127.0.0.10",
			},
		},
		{
			name: "Minimal usage",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt"},
				Format:     "routeros",
				Limit:      0,
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name: "Wrong URLs skipped",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt,foo,google.com/some.txt,,https:// host.com/file"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt"},
				Format:     "routeros",
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name: "URLs unique (duplicates must be removed)",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt,http://foo.com/bar.txt,http://bar.com/baz.asp"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt", "http://bar.com/baz.asp"},
				Format:     "routeros",
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name: "Too long URLs skipped",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt,http://bar.com/baz.asp?x=" + strings.Repeat("x", 232)},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt"},
				Format:     "routeros",
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name:                  "Error when sources url not passed",
			giveURLValues:         url.Values{},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantError:             errors.New("required parameter `sources_urls` was not found"),
		},
		{
			name: "Only wrong values in sources URLs",
			giveURLValues: url.Values{
				"sources_urls": []string{
					"http://foo.com/bar.txt?z=" + strings.Repeat("x", 232) + "," +
						"http://bar.com/baz.asp?x=" + strings.Repeat("x", 232),
				},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantError:             errors.New("empty sources list"),
		},
		{
			name: "Sources limit exceeded",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt,http://bar.com/baz.asp,http://baz.com/blah"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        2,
			wantError:             errors.New("too much sources (only 2 is allowed)"),
		},
		{
			name: "`format` value passing",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"format":       []string{"\t s0me Format_yEah! "},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt"},
				Format:     "\t s0me Format_yEah! ",
				Limit:      0,
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name: "`version` value passing",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"version":      []string{"\t s0me vErSi0n_yEah! "},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt"},
				Format:     "routeros",
				Version:    "\t s0me vErSi0n_yEah! ",
				Limit:      0,
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name: "Excluded host must be unique",
			giveURLValues: url.Values{
				"sources_urls":   []string{"http://foo.com/bar.txt"},
				"excluded_hosts": []string{"foo,bar,foo"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls:    []string{"http://foo.com/bar.txt"},
				Format:        "routeros",
				ExcludedHosts: []string{"foo", "bar"},
				Limit:         0,
				RedirectTo:    "127.0.0.2",
			},
		},
		{
			name: "Empty excluded hosts must be skipped",
			giveURLValues: url.Values{
				"sources_urls":   []string{"http://foo.com/bar.txt"},
				"excluded_hosts": []string{",,,foo,,bar,, ,"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls:    []string{"http://foo.com/bar.txt"},
				Format:        "routeros",
				ExcludedHosts: []string{"foo", "bar", " "},
				Limit:         0,
				RedirectTo:    "127.0.0.2",
			},
		},
		{
			name: "32 excluded host allowed",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"excluded_hosts": []string{"x1,x2,x3,x4,x5,x6,x7,x8,x9,x10,x11,x12,x13,x14,x15,x16,x17,x18,x19,x20," +
					"x21,x22,x23,x24,x25,x26,x27,x28,x29,x30,x31,x32"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        10,
			wantQueryParametersBag: queryParametersBag{
				SourceUrls: []string{"http://foo.com/bar.txt"},
				Format:     "routeros",
				ExcludedHosts: []string{
					"x1", "x2", "x3", "x4", "x5", "x6", "x7", "x8", "x9",
					"x10", "x11", "x12", "x13", "x14", "x15", "x16", "x17", "x18", "x19",
					"x20", "x21", "x22", "x23", "x24", "x25", "x26", "x27", "x28", "x29",
					"x30", "x31", "x32",
				},
				Limit:      0,
				RedirectTo: "127.0.0.2",
			},
		},
		{
			name: "Too many excluded hosts",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"excluded_hosts": []string{"x1,x2,x3,x4,x5,x6,x7,x8,x9,x10,x11,x12,x13,x14,x15,x16,x17,x18,x19,x20," +
					"x21,x22,x23,x24,x25,x26,x27,x28,x29,x30,x31,x32,x33"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        2,
			wantError:             errors.New("too many excluded hosts (more then 32)"),
		},
		{
			name: "Negative 'limit' value",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"limit":        []string{"-1"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        2,
			wantError:             errors.New("wrong `limit` value (cannot be less then 1)"),
		},
		{
			name: "Wrong 'limit' value",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"limit":        []string{"foo"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        2,
			wantError:             errors.New("wrong `limit` value (cannot be converted into integer)"),
		},
		{
			name: "Wrong 'redirect_to' value",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"redirect_to":  []string{"foo"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        2,
			wantError:             errors.New("wrong `redirect_to` value (invalid IP address)"),
		},
		{
			name: "Wrong 'redirect_to' value",
			giveURLValues: url.Values{
				"sources_urls": []string{"http://foo.com/bar.txt"},
				"redirect_to":  []string{"127.0.0.256"},
			},
			giveDefaultRedirectTo: "127.0.0.2",
			giveMaxSources:        2,
			wantError:             errors.New("wrong `redirect_to` value (invalid IP address)"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bag, err := newQueryParametersBag(tt.giveURLValues, tt.giveDefaultRedirectTo, tt.giveMaxSources)

			if tt.wantError != nil && err.Error() != tt.wantError.Error() {
				t.Errorf(`Want error "%v", but got "%v"`, tt.wantError, err)
			}

			if err != nil && tt.wantError == nil {
				t.Errorf(`Error %v returned, but nothing expected`, err)
			}

			if tt.wantError == nil && bag != nil {
				if !reflect.DeepEqual(&tt.wantQueryParametersBag, bag) {
					t.Errorf("Want bag %v, but got %v", tt.wantQueryParametersBag, bag)
				}
			}
		})
	}
}
