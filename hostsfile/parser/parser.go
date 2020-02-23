package parser

import (
	"bufio"
	"errors"
	"io"
	"mikrotik-hosts-parser/hostsfile"
	"net"
	"regexp"
	"strings"
)

// Hosts file parser
type Parser struct {
	hostValidate *regexp.Regexp
}

// NewParser creates new parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse input and return slice of pointers (hosts file entries).
func (p *Parser) Parse(in io.Reader) ([]*hostsfile.Record, error) {
	var (
		result []*hostsfile.Record
		scan   = bufio.NewScanner(in)
	)

	// Read content "line by line"
	for scan.Scan() {
		if entry, err := p.parseRawLine(scan.Text()); err == nil && entry != nil {
			result = append(result, entry)
		}
	}

	if scanErr := scan.Err(); scanErr != nil {
		return result, scanErr
	}

	return result, nil
}

// Validate hostname using regexp.
func (p *Parser) validateHostname(host string) bool {
	// Lazy regexp init
	if p.hostValidate == nil {
		const r string = `(?i)^((-?)(xn--|_)?[a-z0-9-_]{0,61}[a-z0-9-_]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$`
		// @link: https://stackoverflow.com/a/26987741
		p.hostValidate = regexp.MustCompile(r)
	}

	return p.hostValidate.Match([]byte(host))
}

// Parse raw hosts file line into record.
func (p *Parser) parseRawLine(line string) (*hostsfile.Record, error) {
	const delimiter rune = '#'

	// Trim whitespaces
	line = strings.TrimSpace(line)

	// Comment format: `# Any comment text`
	if p.startsWithRune(line, delimiter) {
		return nil, nil
	}

	// Format: `IP_address hostname [host_alias]... #some comment`
	words := strings.Fields(line)

	if len(words) < 2 {
		return nil, errors.New("hosts line parser: wrong line format")
	}

	// first word must be an IP address
	ip := net.ParseIP(words[0])
	if ip == nil {
		return nil, errors.New("hosts line parser: wrong IP address")
	}

	var hosts []string

	for _, host := range words[1:] {
		if p.startsWithRune(host, delimiter) {
			break
		}
		if p.validateHostname(host) {
			hosts = append(hosts, host)
		}
	}

	if len(hosts) == 0 {
		return nil, errors.New("hosts line parser: hosts not found")
	}

	return &hostsfile.Record{IP: ip, Hosts: hosts}, nil
}

// startsWithRune make a check for string starts with passed rune
func (p *Parser) startsWithRune(s string, r rune) bool {
	return len(s) >= 1 && []rune(s)[0] == r
}
