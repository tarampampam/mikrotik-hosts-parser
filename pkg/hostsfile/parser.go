package hostsfile

import (
	"bufio"
	"bytes"
	"io"
	"net/netip"
	"strconv"
	"strings"
	"unsafe"
)

const lineP95SizeBytes = 36

// Record is a hosts file record.
type Record struct {
	IP              string
	Host            string
	AdditionalHosts []string
}

type ParseOption func(*parseOptions)
type parseOptions struct {
	bufferSize   int
	recordsCount int
}

func WithBufferSize(size int) ParseOption {
	return func(opts *parseOptions) {
		opts.bufferSize = size
	}
}

func WithRecordsCount(count int) ParseOption {
	return func(opts *parseOptions) {
		opts.recordsCount = count
	}
}

// Parse input and return slice of records. Result order are same as in source.
//
// You can pass additional options to optimize memory usage and performance.
// For example, if you know the number of records in advance,
// you can use `WithRecordsCount` to preallocate memory for the records slice.
//
// Similarly, you can use `WithBufferSize` to preallocate the internal output buffer
// (used by the parser's strings.Builder).
func Parse(in io.Reader, opts ...ParseOption) ([]Record, error) { //nolint:gocognit,gocyclo,funlen
	var opt parseOptions

	for _, o := range opts {
		o(&opt)
	}

	if opt.recordsCount == 0 {
		if opt.bufferSize != 0 {
			opt.recordsCount = opt.bufferSize / lineP95SizeBytes
		} else {
			opt.recordsCount = 100
		}
	}

	if opt.bufferSize == 0 {
		opt.bufferSize = opt.recordsCount * lineP95SizeBytes
	}

	scan := bufio.NewScanner(in)
	scan.Split(bufio.ScanLines)
	scan.Buffer(make([]byte, 10<<10), 10<<20)

	var (
		ip     bytes.Buffer
		domain bytes.Buffer
	)

	ip.Grow(64)
	domain.Grow(64)

	var wrt writer

	wrt.Init(opt)

scan:
	for scan.Scan() {
		line := scan.Bytes()

		if len(line) <= 5 { //nolint:gomnd
			continue scan // line is too short
		}

		ip.Reset()
		domain.Reset()

		var step int
		const (
			lineStarted = iota
			ipBlock
			betweenBlock
			domainBlock
		)

		for i, r := range line {
			var isLast bool

			if i == len(line)-1 {
				isLast = true
			}

		parseRune:
			switch step {
			case lineStarted, betweenBlock:
				switch {
				case isLast || r == '#': // no records here
					continue scan
				case isSpace(r): // skip spaces
					continue
				default: // block started
					step++ // to ip or to domain

					goto parseRune
				}
			case ipBlock:
				switch {
				case isLast || r == '#':
					continue scan // there is no domain names, skip line
				case isSpace(r):
					// Write ip after then we will find a domain
					step = betweenBlock
				default:
					ip.WriteByte(r)
				}

			case domainBlock:
				if !isSpace(r) && r != '#' {
					domain.WriteByte(r)
				}

				if isSpace(r) || r == '#' || isLast {
					parsedDomain, ok := parseDomain(domain)
					domain.Reset()
					if !ok {
						continue
					}

					if ip.Len() > 0 { // first time sow domain
						parsedIP, ok := parseIP(ip)
						ip.Reset() // reset buffer to avoid double writes
						if !ok {
							continue scan // invalid IP means invalid line
						}

						wrt.NewRecord()
						wrt.Write(parsedIP)
					}

					wrt.Write(parsedDomain)
					step = betweenBlock

					goto parseRune
				}
			}
		} // line parsed
	} // scan finished

	if err := scan.Err(); err != nil {
		return nil, err
	}

	records := wrt.BuildOutput()

	return records, nil
}

type writer struct {
	buf *strings.Builder // buf accumulate all IPs and domains in one string to slice it in a future

	// records count point how many offsets belongs to the record.
	// For example, if record has 3 offsets (IP + Host + AdditionalHost), then recordsCounts will have 3 for this record.
	recordsCounts []int
	offsets       []int // offsets of strings, first one should be 0
	currentRecord int
}

func (w *writer) Init(opt parseOptions) {
	w.recordsCounts = make([]int, 0, opt.recordsCount)
	w.offsets = make([]int, 0, opt.recordsCount*2+1) // IP + Host per record + 1 for 0
	w.buf = new(strings.Builder)
	w.buf.Grow(opt.bufferSize)
	w.currentRecord = -1

	w.offsets = append(w.offsets, 0)
}

func (w *writer) NewRecord() {
	w.currentRecord++
	if len(w.recordsCounts)-1 < w.currentRecord {
		w.recordsCounts = append(w.recordsCounts, 0)
	}
}

func (w *writer) Write(p []byte) {
	w.recordsCounts[w.currentRecord]++
	w.buf.Write(p)
	w.offsets = append(w.offsets, w.buf.Len())
}

func (w *writer) BuildOutput() []Record {
	var (
		result    = make([]Record, len(w.recordsCounts))
		stringMem = w.buf.String() // GC will handle all buffer memory till any link to the string exists
	)

	for i, count := range w.recordsCounts {
		if count > 2 {
			result[i].AdditionalHosts = make([]string, 0, count-2)
		}

		for j := range count {
			rec := stringMem[w.offsets[j]:w.offsets[j+1]]

			switch j {
			case 0:
				result[i].IP = rec
			case 1:
				result[i].Host = rec
			default:
				result[i].AdditionalHosts = append(result[i].AdditionalHosts, rec)
			}
		}

		w.offsets = w.offsets[count:] // shift offsets
	}

	return result
}

func parseIP(in bytes.Buffer) ([]byte, bool) {
	_, err := netip.ParseAddr(unsafe.String(unsafe.SliceData(in.Bytes()), in.Len()))
	if err != nil {
		if longIp := parseLongIP(in.Bytes()); longIp.IsValid() {
			return []byte(longIp.String()), true
		}

		return nil, false
	}

	return in.Bytes(), true
}

func parseDomain(in bytes.Buffer) ([]byte, bool) {
	if validateHostname(in.Bytes()) {
		return in.Bytes(), true
	}

	return nil, false
}

// parseLongIP parses IP address in long format (0 - 4294967295).
func parseLongIP(s []byte) (ip netip.Addr) {
	f, err := strconv.ParseUint(unsafe.String(unsafe.SliceData(s), len(s)), 10, 32)
	if err == nil && f >= 0 && f <= 4294967295 {
		return netip.AddrFrom4([4]byte{
			byte(f >> 24),
			byte(f >> 16),
			byte(f >> 8),
			byte(f),
		})
	}

	return
}

func isSpace(r byte) bool {
	return r == ' ' || r == '\t'
}
