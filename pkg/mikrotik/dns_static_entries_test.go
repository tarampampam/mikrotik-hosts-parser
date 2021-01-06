package mikrotik

import (
	"bytes"
	"github.com/tarampampam/mikrotik-hosts-parser/pkg/mikrotik/dns"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkDNSStaticEntries_Render(b *testing.B) {
	b.ReportAllocs()

	var data DNSStaticEntries

	for i := 0; i < 500; i++ {
		data = append(data, DNSStaticEntry{
			Address:  "0.0.0.0",
			Comment:  "Any text",
			Disabled: true,
			Name:     "www.example.com",
			Regexp:   ".*\\.example\\.com",
			TTL:      "1d",
		})

		data = append(data, DNSStaticEntry{
			Address:  "0.0.0.0",
			Comment:  "Any \\text",
			Disabled: true,
			Name:     "www.example\\.com",
			Regexp:   ".*\\.example\\.com",
			TTL:      "1\\d",
		})
	}

	for n := 0; n < b.N; n++ {
		_, _ = data.Render(ioutil.Discard, "", "")
	}
}

func BenchmarkStaticEntries_RenderOld(b *testing.B) {
	b.ReportAllocs()

	var data dns.StaticEntries

	for i := 0; i < 500; i++ {
		data = append(data, dns.StaticEntry{
			Address:  "0.0.0.0",
			Comment:  "Any text",
			Disabled: true,
			Name:     "www.example.com",
			Regexp:   ".*\\.example\\.com",
			TTL:      "1d",
		})

		data = append(data, dns.StaticEntry{
			Address:  "0.0.0.0",
			Comment:  "Any \\text",
			Disabled: true,
			Name:     "www.example\\.com",
			Regexp:   ".*\\.example\\.com",
			TTL:      "1\\d",
		})
	}

	for n := 0; n < b.N; n++ {
		_, _ = data.Render(ioutil.Discard, &dns.RenderOptions{})
	}
}

func TestDNSStaticEntries_Render(t *testing.T) {
	tests := []struct {
		name        string
		giveEntries DNSStaticEntries
		givePrefix  string
		givePostfix string
		wantResult  string
		wantError   error
	}{
		{
			name:        "empty input",
			giveEntries: DNSStaticEntries{{}},
			wantResult:  "",
		},
		{
			name: "address with comment",
			giveEntries: DNSStaticEntries{{
				Address: "0.0.0.0",
				Comment: "foo comment",
			}},
			wantResult: `address=0.0.0.0 comment="foo comment" disabled=no`,
		},
		{
			name: "two entries with addresses",
			giveEntries: DNSStaticEntries{{
				Address: "0.0.0.0",
			}, {
				Address: "8.8.8.8",
			}},
			wantResult: "address=0.0.0.0 disabled=no\naddress=8.8.8.8 disabled=no",
		},
		{
			name: "two entries (one is empty)",
			giveEntries: DNSStaticEntries{{}, {
				Address: "8.8.8.8",
			}},
			wantResult: "address=8.8.8.8 disabled=no",
		},
		{
			name: "two entries with Prefix and Postfix",
			giveEntries: DNSStaticEntries{{
				Address: "0.0.0.0",
			}, {
				Address: "8.8.8.8",
			}},
			givePrefix:  "foo",
			givePostfix: "bar",
			wantResult:  "foo address=0.0.0.0 disabled=no bar\nfoo address=8.8.8.8 disabled=no bar",
		},
		{
			name: "entry with all fields",
			giveEntries: DNSStaticEntries{{
				Address:  "1.2.3.4",
				Comment:  "foo comment",
				Disabled: true,
				Name:     "Bar name",
				Regexp:   `.*\.example\.com`,
				TTL:      "1d",
			}},
			wantResult: `address=1.2.3.4 comment="foo comment" disabled=yes name="Bar name" regexp=".*\.example\.com" ttl="1d"`,
		},
		{
			name: "regular use-case with address, name and comment",
			giveEntries: DNSStaticEntries{{
				Address: "1.2.3.4",
				Comment: "Foo comment",
				Name:    "Foo entry",
			}, {
				Address: "4.3.2.1",
				Comment: "Bar comment",
				Name:    "Bar entry",
			}},
			wantResult: `address=1.2.3.4 comment="Foo comment" disabled=no name="Foo entry"` + "\n" +
				`address=4.3.2.1 comment="Bar comment" disabled=no name="Bar entry"`,
		},
		{
			name: "Entry with all fields with unescaped values",
			giveEntries: DNSStaticEntries{{
				Address:  "1.2.3.4",
				Comment:  `foo \"bar\" "baz"`,
				Disabled: true,
				Name:     ` "'blah`,
				TTL:      "1d",
			}},
			wantResult: `address=1.2.3.4 comment="foo \"bar\" \"baz\"" disabled=yes name=" \"'blah" ttl="1d"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l, err := tt.giveEntries.Render(&buf, tt.givePrefix, tt.givePostfix)

			assert.Equal(t, len(tt.wantResult), l)

			if tt.wantError != nil {
				assert.EqualError(t, err, tt.wantError.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantResult, buf.String())
		})
	}
}
