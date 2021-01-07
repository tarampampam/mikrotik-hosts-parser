package mikrotik

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkDNSStaticEntries_Render(b *testing.B) {
	b.ReportAllocs()

	var data DNSStaticEntries

	for i := 0; i < 1000; i++ {
		data = append(data, DNSStaticEntry{
			Address:  "0.0.0.0",
			Comment:  "Any text",
			Disabled: true,
			Name:     "www.example.com",
			Regexp:   ".*\\.example\\.com",
			TTL:      "1d",
		})
	}

	for n := 0; n < b.N; n++ {
		_, _ = data.Render(ioutil.Discard)
	}
}

func TestDNSStaticEntries_Render(t *testing.T) {
	tests := []struct {
		name        string
		giveEntries DNSStaticEntries
		giveOptions RenderingOptions
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
				Name:    "foo.com",
				Comment: "foo comment",
			}},
			wantResult: `address=0.0.0.0 comment="foo comment" disabled=no name="foo.com"`,
		},
		{
			name: "two entries with addresses",
			giveEntries: DNSStaticEntries{{
				Address: "0.0.0.0",
				Name:    "foo.com",
			}, {
				Address: "8.8.8.8",
				Name:    "bar.com",
			}},
			wantResult: "address=0.0.0.0 disabled=no name=\"foo.com\"\naddress=8.8.8.8 disabled=no name=\"bar.com\"",
		},
		{
			name: "two entries (one is empty)",
			giveEntries: DNSStaticEntries{{}, {
				Address: "8.8.8.8",
				Name:    "foo.com",
			}},
			wantResult: "address=8.8.8.8 disabled=no name=\"foo.com\"",
		},
		{
			name: "two entries with Prefix and Postfix",
			giveEntries: DNSStaticEntries{{
				Address: "0.0.0.0",
				Name:    "foo.com",
			}, {
				Address: "8.8.8.8",
				Name:    "bar.com",
			}},
			giveOptions: RenderingOptions{
				Prefix:  "foo",
				Postfix: "bar",
			},
			wantResult: "foo address=0.0.0.0 disabled=no name=\"foo.com\" bar\nfoo address=8.8.8.8 disabled=no name=\"bar.com\" bar",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			l, err := tt.giveEntries.Render(&buf, tt.giveOptions)

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
