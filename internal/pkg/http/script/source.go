package script

import (
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/config"

	ver "github.com/tarampampam/mikrotik-hosts-parser/internal/pkg/version"
	"github.com/tarampampam/mikrotik-hosts-parser/pkg/hostsfile"
	"github.com/tarampampam/mikrotik-hosts-parser/pkg/mikrotik"

	"github.com/tarampampam/go-filecache"
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
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// initialize default cache pool
		initDefaultCachePool(serveSettings.Cache.File.DirPath, false)

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
			"cache lifetime: "+strconv.Itoa(int(serveSettings.Cache.LifetimeSec))+" seconds",
		)

		sourceResponsesChannel := make(chan *sourceResponse) // channel for source responses

		// fetch sources async and write responses into channel
		for _, sourceURL := range queryParameters.SourceUrls {
			go writeSourceResponse(
				sourceResponsesChannel,
				sourceURL,
				int(serveSettings.RouterScript.MaxSourceSizeBytes),
				int(serveSettings.Cache.LifetimeSec),
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
func writeSourceResponse(channel chan *sourceResponse, sourceURL string, maxLength, cacheLifetimeSec int) {
	var (
		result    = &sourceResponse{URL: sourceURL}
		cacheItem filecache.CacheItem
	)
	// if cache missed
	if cached := defaultCachePool.GetItem(sourceURL); !cached.IsHit() {
		// do request
		response, fetchError := defaultHTTPClient.FetchSourceContent(sourceURL, maxLength)
		result.Error = fetchError
		if response != nil {
			// and write response content into cache
			cacheItem, _ = defaultCachePool.Put(
				sourceURL,
				response.Body,
				time.Now().Add(time.Second*time.Duration(cacheLifetimeSec)),
			)
			_ = response.Body.Close()
		}
	} else {
		result.CacheIsHit = true
	}
	// extract cached item from cache pool (if was missed previously)
	if cacheItem == nil {
		cacheItem = defaultCachePool.GetItem(sourceURL)
	}
	// read from cache item using pipe
	var pipeReader, pipeWriter = io.Pipe()
	go func() {
		defer func() { _ = pipeWriter.Close() }()
		_ = cacheItem.Get(pipeWriter)
	}()
	result.Content = pipeReader
	if expiresAt := cacheItem.ExpiresAt(); expiresAt != nil {
		result.CacheExpiredAfterSec = int(expiresAt.Unix() - time.Now().Unix())
	}
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
