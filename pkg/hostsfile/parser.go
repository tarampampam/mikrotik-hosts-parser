package hostsfile

import (
	"bufio"
	"errors"
	"io"
	"net"
	"regexp"
	"strings"
)

// hostnameValidationRegex is regular expression for hostname validation, link <https://stackoverflow.com/a/26987741>
var hostnameValidationRegex = regexp.MustCompile(`(?i)^((-?)(xn--|_)?[a-z0-9-_]{0,61}[a-z0-9-_]\.)*(xn--)?([a-z0-9][a-z0-9\-]{0,60}|[a-z0-9-]{1,30}\.[a-z]{2,})$`) //nolint:lll

// Parse input and return slice of records. Result order are same as in source.
func Parse(in io.Reader) ([]Record, error) {
	var (
		result []Record
		scan   = bufio.NewScanner(in)
	)

	// read content "line by line"
	for scan.Scan() {
		if entry, err := parseRawLine(scan.Text()); err == nil && entry != nil {
			result = append(result, *entry)
		}
	}

	if err := scan.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// parseRawLine converts raw hosts file line into Record. nil can be returned without an error (line is empty or
// comment).
// Line format: `IP_address hostname [host_alias]... #some comment`).
func parseRawLine(line string) (*Record, error) {
	if len(line) <= 5 {
		return nil, errors.New("line is too short")
	}

	// skip any lines, that looks like comments in format: `# Any comment text`
	if strings.HasPrefix(line, "#") {
		return nil, nil
	}

	words := strings.Fields(line)

	if len(words) < 2 {
		return nil, errors.New("wrong line format")
	}

	// first word must be an IP address
	ip := net.ParseIP(words[0])
	if ip == nil {
		return nil, errors.New("wrong IP address")
	}

	// map is required for easy duplicates "removal"
	hostsMap := make(map[string]struct{}, len(words)-1)

	for i := 1; i < len(words); i++ {
		if strings.HasPrefix(words[i], "#") {
			break
		}

		if hostnameValidationRegex.MatchString(words[i]) {
			hostsMap[words[i]] = struct{}{}
		}
	}

	if len(hostsMap) == 0 {
		return nil, errors.New("hosts was not found")
	}

	// convert map into slice of strings
	hosts := make([]string, 0, len(hostsMap))
	for host := range hostsMap {
		hosts = append(hosts, host)
	}

	return &Record{IP: ip, Hosts: hosts}, nil
}
