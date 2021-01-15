// Package generate contains RouterOS script generation handler.
package generate

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tarampampam/mikrotik-hosts-parser/v4/internal/pkg/cache"
	"github.com/tarampampam/mikrotik-hosts-parser/v4/internal/pkg/config"
	"github.com/tarampampam/mikrotik-hosts-parser/v4/internal/pkg/version"
	"github.com/tarampampam/mikrotik-hosts-parser/v4/pkg/hostsfile"
	"github.com/tarampampam/mikrotik-hosts-parser/v4/pkg/mikrotik"
	"go.uber.org/zap"
)

type handler struct {
	ctx    context.Context
	log    *zap.Logger
	cacher cache.Cacher
	cfg    *config.Config

	defaultRedirectIP net.IP

	httpClient interface {
		Do(*http.Request) (*http.Response, error)
	}
}

const (
	httpClientTimeout      = time.Second * 10
	httpClientMaxRedirects = 2
	formatRouterOS         = "routeros" //nolint:misspell
)

// NewHandler creates RouterOS script generation handler.
func NewHandler(ctx context.Context, log *zap.Logger, cacher cache.Cacher, cfg *config.Config) (http.Handler, error) {
	if containsIllegalSymbols(cfg.RouterScript.Comment) {
		return nil, errors.New("wrong config: script comment contains illegal symbols")
	}

	if cfg.RouterScript.MaxSourcesCount <= 0 {
		return nil, errors.New("wrong config: max sources count")
	}

	checkRedirectFn := func(req *http.Request, via []*http.Request) error {
		if len(via) >= httpClientMaxRedirects {
			return errors.New("request: too many (2) redirects")
		}

		return nil
	}

	var h = &handler{
		ctx:    ctx,
		log:    log,
		cacher: cacher,
		cfg:    cfg,

		httpClient: &http.Client{Timeout: httpClientTimeout, CheckRedirect: checkRedirectFn},
	}

	if ip := net.ParseIP(cfg.RouterScript.Redirect.Address); ip != nil {
		h.defaultRedirectIP = ip
	} else {
		h.defaultRedirectIP = net.IPv4(127, 0, 0, 1) //nolint:gomnd // config contains wrong value
	}

	return h, nil
}

type hostsFileData struct {
	url      string
	records  []hostsfile.Record
	cacheHit bool
	cacheTTL time.Duration
	err      error
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) { //nolint:funlen,gocognit,gocyclo
	params := newReqParams(h.defaultRedirectIP)

	if r == nil || r.URL == nil {
		w.WriteHeader(http.StatusBadRequest)
		h.writeComment(w, "Empty request or query parameters")

		return
	}

	if err := params.fromValues(r.URL.Query()); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.writeComment(w, "Query parameters error: "+err.Error())

		return
	}

	if err := params.validate(h.cfg.RouterScript.MaxSourcesCount); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.writeComment(w, "Query parameters validation failed: "+err.Error())

		return
	}

	if format := params.format; format != formatRouterOS {
		w.WriteHeader(http.StatusBadRequest)
		h.writeComment(w, fmt.Sprintf("Unsupported format [%s] requested", format))

		return
	}

	// write script header
	h.writeComment(w,
		"Script generated at "+time.Now().Format("2006-01-02 15:04:05"),
		"Generator version: "+version.Version(),
		fmt.Sprintf("Limit: %d", params.limit),
		fmt.Sprintf("Cache lifetime: %s", h.cacher.TTL().Round(time.Second)),
		"Format: "+params.format,
		"Redirect to: "+params.redirect.String(),
		"Sources list:",
	)

	for i := 0; i < len(params.sources); i++ {
		h.writeComment(w, fmt.Sprintf(" - <%s>", params.sources[i]))
	}

	if len(params.excluded) > 0 {
		h.writeComment(w, "Excluded hosts:")

		for i := 0; i < len(params.excluded); i++ {
			h.writeComment(w, fmt.Sprintf(" - %s", params.excluded[i]))
		}
	}

	var (
		hostsDataCh       = make(chan hostsFileData, len(params.sources))
		hostsRecordsCount uint32 // atomic usage only, used for hosts list pre-allocation
		wg                sync.WaitGroup
	)

	wg.Add(len(params.sources))

	// fetch hosts files content and parse them
	for i := 0; i < len(params.sources); i++ {
		go func(ch chan<- hostsFileData, url string) {
			defer wg.Done()

			if hit, data, ttl, err := h.cacher.Get(url); hit && err == nil {
				records, parsingErr := hostsfile.Parse(bytes.NewReader(data))
				if parsingErr == nil {
					atomic.AddUint32(&hostsRecordsCount, uint32(len(records)))
					ch <- hostsFileData{url: url, records: records, cacheHit: hit, cacheTTL: ttl}

					return
				}

				ch <- hostsFileData{url: url, cacheHit: hit, err: parsingErr}

				return
			}

			data, srcErr := h.fetchRemoteSource(url)
			if srcErr != nil {
				h.log.Warn("remote source fetching failed", zap.Error(srcErr), zap.String("url", url))
				ch <- hostsFileData{url: url, err: srcErr}

				return
			}

			if err := h.cacher.Put(url, data.Bytes()); err != nil {
				h.log.Error("cache writing error", zap.Error(err), zap.String("url", url))
				ch <- hostsFileData{url: url, err: err}

				return
			}

			if records, err := hostsfile.Parse(data); err == nil {
				atomic.AddUint32(&hostsRecordsCount, uint32(len(records)))
				ch <- hostsFileData{url: url, records: records, cacheTTL: h.cacher.TTL()}
			} else {
				ch <- hostsFileData{url: url, err: err}
			}
		}(hostsDataCh, params.sources[i])
	}

	wg.Wait()
	close(hostsDataCh)

	if err := h.ctx.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		h.writeComment(w, "Context error: "+err.Error())

		return
	}

	// burn excludes map for fastest checking
	var excludes = make(map[string]struct{}, len(params.excluded))
	for i := 0; i < len(params.excluded); i++ {
		excludes[params.excluded[i]] = struct{}{}
	}

	// calculate results map size for pre-allocation
	var size uint32
	if params.limit > 0 {
		size = params.limit
	} else {
		size = atomic.LoadUint32(&hostsRecordsCount)
	}

	var hostNames, limit = make(map[string]struct{}, size), int(size)

	// read parsed hosts files content from channel
	for i := 0; i < len(params.sources); i++ {
		data := <-hostsDataCh

		if data.err != nil {
			h.writeComment(w, fmt.Sprintf("Source <%s> error: %v", data.url, data.err))

			continue
		}

		if data.cacheHit {
			h.writeComment(w, fmt.Sprintf("Cache HIT for <%s> (expires after %s)", data.url, data.cacheTTL.Round(time.Second)))
		} else {
			h.writeComment(w, fmt.Sprintf("Cache miss for <%s>", data.url))
		}

	recordsLoop:
		for j := 0; j < len(data.records); j++ { // loop over records inside hosts file
			if name := data.records[j].Host; name != "" {
				if len(hostNames) >= limit { // hostnames limit has been reached
					break recordsLoop
				}

				if !containsIllegalSymbols(name) {
					if _, ok := excludes[name]; !ok { // is in excludes list?
						hostNames[name] = struct{}{} // append
					}
				}
			}

			if len(data.records[j].AdditionalHosts) > 0 { //nolint:nestif
				for k := 0; k < len(data.records[j].AdditionalHosts); k++ { // loop over additional hostnames
					if len(hostNames) >= limit { // hostnames limit has been reached
						break recordsLoop
					}

					name := data.records[j].AdditionalHosts[k]

					if _, ok := excludes[name]; ok { // is in excludes list?
						continue
					}

					if !containsIllegalSymbols(name) {
						if _, ok := excludes[name]; !ok { // is in excludes list?
							hostNames[name] = struct{}{} // append
						}
					}
				}
			}
		}
	}

	if len(hostNames) == 0 {
		h.writeComment(w, "Script generation failed (empty hosts list)")

		return
	}

	var result, redirectAddr = make(mikrotik.DNSStaticEntries, 0, len(hostNames)), params.redirect.String()
	for hostName := range hostNames {
		result = append(result, mikrotik.DNSStaticEntry{
			Address: redirectAddr,
			Comment: h.cfg.RouterScript.Comment,
			Name:    hostName,
		})
	}

	// make sorting
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	_, _ = w.Write([]byte("\n/ip dns static\n"))
	_, renderingErr := result.Render(w, mikrotik.RenderingOptions{Prefix: "add"})
	_, _ = w.Write([]byte("\n\n"))

	if renderingErr != nil {
		h.writeComment(w, fmt.Sprintf("Script rendering error: %v", renderingErr))
	}

	h.writeComment(w, fmt.Sprintf(
		"Records count: %d (%d records ignored)",
		len(result),
		int(atomic.LoadUint32(&hostsRecordsCount))-len(result),
	))
}

func containsIllegalSymbols(s string) bool {
	return strings.ContainsRune(s, '"') || strings.ContainsRune(s, '\\')
}

func (h *handler) writeComment(w io.Writer, comments ...string) {
	for i := 0; i < len(comments); i++ {
		_, _ = w.Write([]byte("## " + comments[i] + "\n"))
	}
}

func (h *handler) fetchRemoteSource(url string) (*bytes.Buffer, error) {
	req, err := http.NewRequestWithContext(h.ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("wrong response code: %d", resp.StatusCode)
	}

	if ct, allowed := resp.Header.Get("Content-Type"), "text/plain"; !strings.HasPrefix(ct, allowed) {
		return nil, fmt.Errorf("wrong Content-Type response header [%s] (%s* is required)", ct, allowed)
	}

	var buf bytes.Buffer

	const defaultBufCapacity = 64 * 1024 // 64 KiB

	if cl := resp.Header.Get("Content-Length"); cl != "" { //nolint:nestif
		value, parsingErr := strconv.Atoi(cl)
		if parsingErr != nil {
			return nil, errors.New("header Content-Length parsing error: " + parsingErr.Error())
		}

		if max := int(h.cfg.RouterScript.MaxSourceSizeBytes); value >= max {
			return nil, fmt.Errorf("header Content-Length value [%d] is too big (max: %d)", value, max)
		}

		if value > 0 {
			buf.Grow(value)
		} else {
			buf.Grow(defaultBufCapacity)
		}
	} else {
		buf.Grow(defaultBufCapacity)
	}

	if _, readingErr := buf.ReadFrom(resp.Body); readingErr != nil {
		return nil, readingErr
	}

	return &buf, nil
}

type reqParams struct {
	sources  []string
	format   string
	ver      string
	excluded []string
	limit    uint32
	redirect net.IP
}

func newReqParams(redirect net.IP) reqParams {
	return reqParams{
		sources:  make([]string, 0, 8),
		format:   formatRouterOS, // default value
		excluded: make([]string, 0, 16),
		redirect: redirect,
	}
}

func (p *reqParams) fromValues(v url.Values) error { //nolint:funlen,gocognit,gocyclo
	if urls, ok := v["sources_urls"]; ok {
		m := make(map[string]struct{}, 8)

		for i := 0; i < len(urls); i++ {
			for list, j := strings.Split(urls[i], ","), 0; j < len(list); j++ {
				if u, err := url.ParseRequestURI(list[j]); err == nil {
					m[u.String()] = struct{}{}
				}
			}
		}

		for u := range m {
			p.sources = append(p.sources, u)
		}

		sort.Strings(p.sources)
	} else {
		return errors.New("required parameter 'sources_urls' was not found")
	}

	if value, ok := v["format"]; ok { // optional
		if len(value) > 0 {
			p.format = value[0]
		}
	}

	if value, ok := v["version"]; ok { // optional
		if len(value) > 0 {
			p.ver = value[0]
		}
	}

	if hosts, ok := v["excluded_hosts"]; ok { // optional
		m := make(map[string]struct{}, 16)

		for i := 0; i < len(hosts); i++ {
			for list, j := strings.Split(hosts[i], ","), 0; j < len(list); j++ {
				if host := list[j]; host != "" {
					m[strings.Trim(host, " '\"\n\r")] = struct{}{}
				}
			}
		}

		for host := range m {
			p.excluded = append(p.excluded, host)
		}

		sort.Strings(p.excluded)
	}

	if value, ok := v["limit"]; ok { // optional
		if len(value) > 0 {
			if limit, err := strconv.ParseUint(value[0], 10, 32); err == nil && limit > 0 {
				p.limit = uint32(limit)
			} else {
				return errors.New("wrong 'limit' value")
			}
		}
	}

	if value, ok := v["redirect_to"]; ok { // optional
		if len(value) > 0 {
			ip := net.ParseIP(value[0])
			if ip == nil {
				return errors.New("wrong 'redirect_to' value (invalid IP address)")
			}

			p.redirect = ip
		}
	}

	return nil
}

func (p *reqParams) validate(maxSources uint16) error {
	if l := len(p.sources); l == 0 {
		return errors.New("empty sources list")
	} else if l > int(maxSources) {
		return fmt.Errorf("too many sources (only %d is allowed)", maxSources)
	}

	if len(p.excluded) > 32 { //nolint:gomnd
		return errors.New("too many excluded hosts (more then 32)")
	}

	return nil
}
