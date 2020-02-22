package script

import (
	"io"
	"mikrotik-hosts-parser/hostsfile"
	hostsParser "mikrotik-hosts-parser/hostsfile/parser"
	"mikrotik-hosts-parser/mikrotik/dns"
	"mikrotik-hosts-parser/settings/serve"
	ver "mikrotik-hosts-parser/version"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type sourceResponse struct {
	URL     string
	Content io.ReadCloser
	Error   error
}

// RouterOsScriptSourceGenerationHandlerFunc generates RouterOS script source and writes it response.
func RouterOsScriptSourceGenerationHandlerFunc( //nolint:funlen
	serveSettings *serve.Settings,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		queryParameters, queryErr := newQueryParametersBag(
			r.URL.Query(),
			serveSettings.RouterScript.Redirect.Address,
			serveSettings.RouterScript.MaxSources,
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
		)

		sourceResponsesChannel := make(chan sourceResponse) // channel for source responses

		// fetch sources async and write responses into channel
		for _, sourceURL := range queryParameters.SourceUrls {
			go func(sourceURL string, maxLength int) {
				var content io.ReadCloser
				// do request
				response, err := defaultHTTPClient.FetchSourceContent(sourceURL, maxLength) //nolint:bodyclose
				if response != nil {
					content = response.Body // content MUST BE CLOSED (later)!
				}
				// send request result into channel
				sourceResponsesChannel <- sourceResponse{URL: sourceURL, Content: content, Error: err}
			}(sourceURL, serveSettings.RouterScript.MaxSourceSize)
		}

		var (
			parser       = hostsParser.NewParser()
			hostsRecords = make([]*hostsfile.Record, 0) // hosts records stack
		)

		// read source responses and pass it into hosts file parser
		for i := 0; i < len(queryParameters.SourceUrls); i++ {
			// read message from channel
			resp := <-sourceResponsesChannel
			// if response contains error - skip it
			if resp.Error != nil {
				if resp.Content != nil {
					_ = resp.Content.Close()
				}
				comments = append(comments, "Source <"+resp.URL+"> error: "+resp.Error.Error())
				continue
			}
			// parse response content
			records, parseErr := parser.Parse(resp.Content)
			_ = resp.Content.Close()
			if parseErr != nil {
				comments = append(comments, "Source <"+resp.URL+"> error: "+parseErr.Error())
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
			_, renderErr := staticEntries.Render(w, &dns.RenderOptions{
				RenderEntryOptions: dns.RenderEntryOptions{
					Prefix: "add",
				},
				RenderEmpty: false,
			})

			if renderErr != nil {
				_, _ = w.Write([]byte("\n\n## Rendering error: " + renderErr.Error()))
			}
		}
	}
}

// hostsRecordsToStaticEntries converts hosts records into static dns entries
func hostsRecordsToStaticEntries(
	in []*hostsfile.Record,
	excludes []string,
	limit int,
	redirectTo,
	comment string,
) dns.StaticEntries {
	var (
		processedHosts = make(map[string]bool)
		out            = dns.StaticEntries{}
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
				out = append(out, dns.StaticEntry{
					Address: redirectTo,
					Comment: comment,
					Name:    host,
				})
			}
		}
	}

	return out
}
