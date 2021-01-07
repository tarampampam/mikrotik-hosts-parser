package mikrotik

type DNSStaticEntry struct {
	Address  string // IP address (net.IP is not used for allocation avoiding reasons (to string), eg.: 0.0.0.0)
	Comment  string // Short description of the item (eg.: Any text)
	Disabled bool   // Defines whether item is ignored or used (eg.: yes,no)
	Name     string // Host name (eg.: www.example.com)
	Regexp   string // Regular expression (eg.: .*\\.example\\.com)
	TTL      string // Time To Live (eg.: 1d)
}

// Format entry as a text in RouterOS script format.
// Important: keep im mind that any unexpected characters will be formatted as-is (without escaping or filtering).
func (s *DNSStaticEntry) Format(prefix, postfix string) ([]byte, error) {
	buf := make([]byte, 0, len(s.Address)+len(s.Comment)+len(s.Name)+len(s.Regexp)+len(s.TTL)+96)
	err := s.format(&buf, prefix, postfix)

	return buf, err
}

// format documentation: <https://wiki.mikrotik.com/wiki/Manual:IP/DNS#Static_DNS_Entries>
// Important: empty values will NOT be printed. Values escaping are not allowed here (reason - allocation avoiding).
func (s *DNSStaticEntry) format(buf *[]byte, prefix, postfix string) error {
	if s.Address == "" || (s.Name == "" && s.Regexp == "") {
		return ErrEmptyFields
	}

	// write prefix
	if len(prefix) > 0 {
		*buf = append(*buf, prefix+" "...)
	}

	// write "address"
	*buf = append(*buf, `address=`+s.Address...) // quoting ("..") is needed here?

	// write "comment"
	if s.Comment != "" {
		*buf = append(*buf, ` comment="`+s.Comment+`"`...)
	}

	// write "disabled"
	*buf = append(*buf, ` disabled=`+s.boolToString(s.Disabled)...)

	// write "name"
	if s.Name != "" {
		*buf = append(*buf, ` name="`+s.Name+`"`...)
	}

	// write "regexp"
	if s.Regexp != "" {
		*buf = append(*buf, ` regexp="`+s.Regexp+`"`...)
	}

	// write "ttl"
	if s.TTL != "" {
		*buf = append(*buf, ` ttl="`+s.TTL+`"`...)
	}

	// write entry Postfix
	if len(postfix) > 0 {
		*buf = append(*buf, " "+postfix...)
	}

	return nil
}

func (DNSStaticEntry) boolToString(b bool) string {
	if b {
		return "yes"
	}

	return "no"
}
