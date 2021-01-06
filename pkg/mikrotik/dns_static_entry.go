package mikrotik

import "strings"

type DNSStaticEntry struct {
	Address  string // IP address (net.IP is not used for allocation avoiding reasons (to string), eg.: 0.0.0.0)
	Comment  string // Short description of the item (eg.: Any text)
	Disabled bool   // Defines whether item is ignored or used (eg.: yes,no)
	Name     string // Host name (eg.: www.example.com)
	Regexp   string // Regular expression (eg.: .*\\.example\\.com)
	TTL      string // Time To Live (eg.: 1d)
}

// Format entry as a text in RouterOS script format.
// Important: empty values will not be printed.
func (s *DNSStaticEntry) Format(prefix, postfix string) []byte {
	b := make([]byte, 0, len(s.Address)+len(s.Comment)+len(s.Name)+len(s.Regexp)+len(s.TTL)+96)
	s.format(&b, prefix, postfix)
	return b
}

// format documentation: <https://wiki.mikrotik.com/wiki/Manual:IP/DNS#Static_DNS_Entries>
func (s *DNSStaticEntry) format(buf *[]byte, prefix, postfix string) {
	// write prefix
	if len(prefix) > 0 {
		*buf = append(*buf, prefix+" "...)
	}

	// write "address"
	*buf = append(*buf, `address=`+s.escapeString(s.Address)...)

	// write "comment"
	if s.Comment != "" {
		*buf = append(*buf, ` comment="`+s.escapeString(s.Comment)+`"`...)
	}

	// write "disabled"
	*buf = append(*buf, ` disabled=`+s.boolToString(s.Disabled)...)

	// write "name"
	if s.Name != "" {
		*buf = append(*buf, ` name="`+s.escapeString(s.Name)+`"`...)
	}

	// write "regexp"
	if s.Regexp != "" {
		*buf = append(*buf, ` regexp="`+s.Regexp+`"`...)
	}

	// write "ttl"
	if s.TTL != "" {
		*buf = append(*buf, ` ttl="`+s.escapeString(s.TTL)+`"`...)
	}

	// write entry Postfix
	if len(postfix) > 0 {
		*buf = append(*buf, " "+postfix...)
	}
}

func (DNSStaticEntry) boolToString(b bool) string {
	if b {
		return "yes"
	}

	return "no"
}

func (DNSStaticEntry) escapeString(s string) string { // TODO performance can be better here
	for _, char := range s { // strings.ContainsRune(s, '\\') || strings.ContainsRune(s, '"')
		if char == '"' || char == '\\' {
			return strings.Replace(strings.Replace(s, `\`, ``, -1), `"`, `\"`, -1)
		}
	}

	return s
}
