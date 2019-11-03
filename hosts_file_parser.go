package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"regexp"
	"strings"
)

type (
	// Hosts file record
	HostsFileRecord struct {
		IP    net.IP
		Hosts []string
	}

	// Hosts file parser
	HostsFileParser struct {
		hostValidate *regexp.Regexp
	}
)

// Parse input and return slice of pointers (hosts file entries).
func (p *HostsFileParser) Parse(in io.Reader) ([]*HostsFileRecord, error) {
	var (
		result []*HostsFileRecord
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
func (p *HostsFileParser) validateHostname(host string) bool {
	// Lazy regexp init
	if p.hostValidate == nil {
		// @link: https://stackoverflow.com/a/26987741
		p.hostValidate = regexp.MustCompile(`(?i)^((-?)(xn--|_)?[a-z0-9-_]{0,61}[a-z0-9-_]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$`)
	}

	return p.hostValidate.Match([]byte(host))
}

// Parse raw hosts file line into record.
func (p *HostsFileParser) parseRawLine(line string) (*HostsFileRecord, error) {
	const delimiter rune = '#'

	// Trim whitespaces
	line = strings.TrimSpace(line)

	// Comment format: `# Any comment text`
	if len(line) >= 1 && []rune(line)[0] == delimiter {
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
		//if strings.IndexRune(host, delimiter) == 0 {
		if len(host) >= 1 && []rune(host)[0] == delimiter {
			break
		}
		if p.validateHostname(host) {
			hosts = append(hosts, host)
		}
	}

	if len(hosts) == 0 {
		return nil, errors.New("hosts line parser: hosts not found")
	}

	return &HostsFileRecord{IP: ip, Hosts: hosts}, nil
}
