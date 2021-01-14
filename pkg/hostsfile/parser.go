package hostsfile

import (
	"bufio"
	"bytes"
	"io"
	"net"
)

// Hostname validator generation (execute in linux shell) using <https://gitlab.com/opennota/re2dfa>:
//	$ cd ./pkg/hostsfile
//	$ docker run --rm -ti -v $(pwd):/rootfs:rw -w /rootfs golang:1.15-buster
//	$ go get -u gitlab.com/opennota/re2dfa
//	$ re2dfa -o hostname_validator.go \
//	'(?i)^((-?)(xn--|_)?[a-z0-9-_]{0,61}[a-z0-9-_]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$' \
//	hostsfile.validateHostname []byte
//	$ exit
//	$ sudo chown "$(id -u):$(id -g)" ./hostname_validator.go

type wordFlag uint8

func (f wordFlag) HasFlag(flag wordFlag) bool { return f&flag != 0 }
func (f *wordFlag) AddFlag(flag wordFlag)     { *f |= flag }
func (f *wordFlag) ClearFlag(flag wordFlag)   { *f &= ^flag }
func (f *wordFlag) Reset()                    { *f = wordFlag(0) }

const (
	wordEnded wordFlag = 1 << iota
	wordWithDot
	wordWithColon
)

type word struct {
	buf    bytes.Buffer
	count  uint
	flag   wordFlag
	isLast bool
}

func (w *word) Reset() {
	w.count = 0
	w.flag.Reset()
	w.isLast = false
	w.buf.Reset()
}

// Parse input and return slice of records. Result order are same as in source.
func Parse(in io.Reader) ([]Record, error) { //nolint:funlen,gocognit,gocyclo
	var (
		result    = make([]Record, 0, 5)
		scan      = bufio.NewScanner(in)
		w         word
		hostnames = make([]string, 0, 3)
		ip        bytes.Buffer
	)

	w.buf.Grow(32) //nolint:gomnd
	ip.Grow(7)     //nolint:gomnd

scan: // read content "line by line"
	for scan.Scan() {
		line := scan.Bytes()

		if len(line) <= 5 { //nolint:gomnd
			continue scan // line is too short
		}

		if line[0] == '#' {
			continue scan // skip any lines, that looks like comments in format: `# Any comment text`
		}

		w.Reset()
		ip.Reset()
		if len(hostnames) > 0 {
			hostnames = hostnames[:0]
		}

		for i, ll := 0, len(line); i < ll && !w.isLast; i++ { // loop over line runes
			if char := line[i]; char != ' ' && char != '\t' {
				if char == '.' {
					w.flag.AddFlag(wordWithDot)
				} else if char == ':' {
					w.flag.AddFlag(wordWithColon)
				}

				w.buf.WriteByte(char)
				w.flag.ClearFlag(wordEnded)
			} else {
				w.flag.AddFlag(wordEnded)
			}

			if w.flag.HasFlag(wordEnded) || i == ll-1 { //nolint:nestif // word filled completely
				if w.buf.Len() == 0 {
					continue // skip any empty words
				}

				w.count++

				if w.count == 1 && w.buf.Bytes()[0] == '#' {
					continue scan // skip if first word starts with comment char
				}

				if w.count == 1 {
					if (w.flag.HasFlag(wordWithDot) && validateIPv4(w.buf.Bytes())) ||
						(w.flag.HasFlag(wordWithColon) && net.ParseIP(w.buf.String()) != nil) {
						ip.Write(w.buf.Bytes())
					}
				} else {
					if w.buf.Bytes()[0] == '#' { // comment at the end of line
						w.isLast = true
					} else if ip.Len() > 0 && validateHostname(w.buf.Bytes()) > 0 {
						hostnames = append(hostnames, w.buf.String()) // +1 memory allocation here
					}
				}

				w.buf.Reset()
				w.flag.Reset()
			}
		}

		if ip.Len() > 0 && len(hostnames) > 0 {
			rec := Record{IP: ip.String(), Host: hostnames[0]} // +1 memory allocation here

			if l := len(hostnames); l > 1 {
				rec.AdditionalHosts = make([]string, 0, l-1) // +1 memory allocation here (but not for each record)
				rec.AdditionalHosts = append(rec.AdditionalHosts, hostnames[1:]...)
			}

			result = append(result, rec)
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// validateIPv4 address (d.d.d.d).
func validateIPv4(s []byte) bool {
	var p [net.IPv4len]byte

	for i := 0; i < net.IPv4len; i++ {
		if len(s) == 0 {
			return false // missing octets
		}

		if i > 0 {
			if s[0] != '.' {
				return false
			}

			s = s[1:]
		}

		n, c, ok := dtoi(s)
		if !ok || n > 0xFF {
			return false
		}

		s = s[c:]
		p[i] = byte(n)
	}

	return len(s) == 0
}

// dtoi converts decimal to integer. Returns number, characters consumed, success.
func dtoi(s []byte) (n int, i int, ok bool) {
	const big = 0xFFFFFF

	n = 0
	for i = 0; i < len(s) && '0' <= s[i] && s[i] <= '9'; i++ {
		n = n*10 + int(s[i]-'0') //nolint:gomnd
		if n >= big {
			return big, i, false
		}
	}

	if i == 0 {
		return 0, 0, false
	}

	return n, i, true
}
