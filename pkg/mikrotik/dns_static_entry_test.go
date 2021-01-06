package mikrotik

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkDNSStaticEntry_Format(b *testing.B) {
	b.ReportAllocs()

	s := DNSStaticEntry{
		Address:  "0.0.0.0",
		Comment:  "Any text",
		Disabled: true,
		Name:     "www.example.com",
		Regexp:   ".*\\.example\\.com",
		TTL:      "1d",
	}

	for n := 0; n < b.N; n++ {
		_ = s.Format("foo", "bar")
	}
}

func BenchmarkDNSStaticEntry_FormatWithEscaping(b *testing.B) {
	b.ReportAllocs()

	s := DNSStaticEntry{
		Address:  "0.0.0.0",
		Comment:  "Any\\ text",
		Disabled: true,
		Name:     "www.example\\.com",
		Regexp:   ".*\\.example\\.com",
		TTL:      "1\\d",
	}

	for n := 0; n < b.N; n++ {
		_ = s.Format("foo", "bar")
	}
}

func TestDNSStaticEntry_Format(t *testing.T) {
	cases := []struct {
		name        string
		giveEntry   DNSStaticEntry
		givePrefix  string
		givePostfix string
		wantString  string
	}{
		{
			name: "regular usage",
			giveEntry: DNSStaticEntry{
				Address:  "0.0.0.0",
				Comment:  "Any text",
				Disabled: true,
				Name:     "www.example.com",
				Regexp:   `.*\.example\.com`,
				TTL:      "1d",
			},
			givePrefix:  "foo",
			givePostfix: "bar",
			wantString:  `foo address=0.0.0.0 comment="Any text" disabled=yes name="www.example.com" regexp=".*\.example\.com" ttl="1d" bar`, //nolint:lll
		},
		{
			name: "minimal usage",
			giveEntry: DNSStaticEntry{
				Address:  "0.0.0.0",
			},
			wantString:  `address=0.0.0.0 disabled=no`,
		},
		{
			name: "with escaping",
			giveEntry: DNSStaticEntry{
				Address:  "127.0.0.1",
				Comment:  "Any\\ text",
				Name:     "www.example\\.com",
				Regexp:   `.*\.example\.com`,
				TTL:      "1\\d",
			},
			wantString:  `address=127.0.0.1 comment="Any text" disabled=no name="www.example.com" regexp=".*\.example\.com" ttl="1d"`, //nolint:lll
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantString, string(tt.giveEntry.Format(tt.givePrefix, tt.givePostfix)))
		})
	}
}
