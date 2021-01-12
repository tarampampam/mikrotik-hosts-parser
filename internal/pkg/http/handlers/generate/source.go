package generate

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/cache"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"

	ver "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/version"
	"github.com/tarampampam/mikrotik-hosts-parser/pkg/hostsfile"
	"github.com/tarampampam/mikrotik-hosts-parser/pkg/mikrotik"
)

type sourceResponse struct {
	URL                  string
	Content              io.ReadCloser
	Error                error
	CacheIsHit           bool
	CacheExpiredAfterSec int
}

// RouterOsScriptSourceGenerationHandlerFunc generates RouterOS script source and writes it response.
func RouterOsScriptSourceGenerationHandlerFunc( //nolint:funlen,gocyclo
	serveSettings *config.Config,
	cacher cache.Cacher,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParameters, queryErr := newQueryParametersBag(
			r.URL.Query(),
			serveSettings.RouterScript.Redirect.Address,
			int(serveSettings.RouterScript.MaxSourcesCount),
		)

		// Validate query parameters parsing
		if queryErr != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("## Query parameters error: " + queryErr.Error()))

			return
		}

		comments := make([]string, 0) // strings slice for storing processing comments (e.g. info messages, errors, etc)

		// append basic information
		comments = append(comments,
			"Script generated at "+time.Now().Format("2006-01-02 15:04:05"),
			"Generator version: "+ver.Version(),
			"",
			"Sources list: <"+strings.Join(queryParameters.SourceUrls, ">, <")+">",
			"Excluded hosts: '"+strings.Join(queryParameters.ExcludedHosts, "', '")+"'",
			"Limit: "+strconv.Itoa(queryParameters.Limit),
			"cache lifetime: "+strconv.Itoa(int(cacher.TTL()))+" seconds",
		)

		sourceResponsesChannel := make(chan *sourceResponse) // channel for source responses

		// fetch sources async and write responses into channel
		for _, sourceURL := range queryParameters.SourceUrls {
			go writeSourceResponse(
				cacher,
				sourceResponsesChannel,
				sourceURL,
				int(serveSettings.RouterScript.MaxSourceSizeBytes),
				int(cacher.TTL()),
			)
		}

		hostsRecords := make([]hostsfile.Record, 0) // hosts records stack

		// read source responses and pass it into hosts file parser
		for i := 0; i < len(queryParameters.SourceUrls); i++ {
			// read message from channel
			resp := <-sourceResponsesChannel
			if resp.CacheIsHit {
				comments = append(comments, "cache HIT for <"+resp.URL+"> "+
					"(expires after "+strconv.Itoa(resp.CacheExpiredAfterSec)+" sec.)")
			} else {
				comments = append(comments, "cache miss for <"+resp.URL+">")
			}
			// if response contains error - skip it
			if resp.Error != nil {
				if resp.Content != nil {
					_ = resp.Content.Close()
				}
				comments = append(comments, "source <"+resp.URL+"> error: "+resp.Error.Error())
				continue
			}
			// parse response content
			records, parseErr := hostsfile.Parse(resp.Content)
			_ = resp.Content.Close()
			if parseErr != nil {
				comments = append(comments, "source <"+resp.URL+"> error: "+parseErr.Error())
			}
			// and append results into hosts records stack
			hostsRecords = append(hostsRecords, records...)
		}

		// close responses channels after all
		close(sourceResponsesChannel)

		// convert hosts records into static mikrotik dns entries
		staticEntries := hostsRecordsToStaticEntries(
			hostsRecords,
			queryParameters.ExcludedHosts,
			queryParameters.Limit,
			queryParameters.RedirectTo,
			serveSettings.RouterScript.Comment,
		)

		// write processing comments
		for _, comment := range comments {
			buf := make([]byte, 0)
			if comment == "" {
				buf = append(buf, "\n"...)
			} else {
				buf = append(buf, "## "+comment+"\n"...)
			}
			_, _ = w.Write(buf)
		}

		// render result script source
		if len(staticEntries) > 0 {
			_, _ = w.Write([]byte("\n/ip dns static\n"))
			_, renderErr := staticEntries.Render(w, mikrotik.RenderingOptions{
				Prefix: "add",
			})

			_, _ = w.Write([]byte("\n\n## Records count: " + strconv.Itoa(len(staticEntries))))
			if renderErr != nil {
				_, _ = w.Write([]byte("\n\n## Rendering error: " + renderErr.Error()))
			}
		}
	}
}

// writeSourceResponse writes source response into channel (content can be fetched from cache)
func writeSourceResponse(cacher cache.Cacher, channel chan *sourceResponse, sourceURL string, maxLength, cacheLifetimeSec int) {
	var result = &sourceResponse{URL: sourceURL}

	// if cache missed
	if hit, _, _, _ := cacher.Get(sourceURL); !hit {
		// do request
		response, fetchError := defaultHTTPClient.FetchSourceContent(sourceURL, maxLength)
		result.Error = fetchError
		if response != nil {
			bodyBytes, _ := ioutil.ReadAll(response.Body)
			_ = response.Body.Close()

			// and write response content into cache
			_ = cacher.Put(sourceURL, bodyBytes)
		}
	} else {
		result.CacheIsHit = true
	}

	_, data, ttl, _ := cacher.Get(sourceURL)

	result.CacheExpiredAfterSec = int(ttl.Seconds())
	result.Content = ioutil.NopCloser(bytes.NewReader(data))

	channel <- result
}

// hostsRecordsToStaticEntries converts hosts records into static dns entries
func hostsRecordsToStaticEntries(
	in []hostsfile.Record,
	excludes []string,
	limit int,
	redirectTo,
	comment string,
) mikrotik.DNSStaticEntries {
	var (
		processedHosts = make(map[string]bool)
		out            = mikrotik.DNSStaticEntries{}
	)

	// put hosts for excluding into processed hosts map for skipping in future
	for _, host := range excludes {
		processedHosts[host] = true
	}

	// loop over all passed hosts file records
records:
	for _, record := range in {
		// iterate hosts in record
		for _, host := range record.Hosts {
			// maximal hosts checking
			if limit > 0 && len(out) >= limit {
				break records
			}
			// verification that host was not processed previously
			if _, ok := processedHosts[host]; !ok {
				// set "was processed" flag in hosts map
				processedHosts[host] = true
				// add new static entry into result
				out = append(out, mikrotik.DNSStaticEntry{
					Address: redirectTo,
					Comment: comment,
					Name:    host,
				})
			}
		}
	}

	// make sorting
	sort.Slice(out[:], func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}
