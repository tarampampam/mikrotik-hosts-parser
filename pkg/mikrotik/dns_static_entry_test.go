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
		_, _ = s.Format("foo", "bar")
	}
}

func TestDNSStaticEntry_Format(t *testing.T) {
	cases := []struct {
		name        string
		giveEntry   DNSStaticEntry
		givePrefix  string
		givePostfix string
		wantString  string
		wantError   error
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
				Address: "0.0.0.0",
				Name:    "foo.com",
			},
			wantString: `address=0.0.0.0 disabled=no name="foo.com"`,
		},
		{
			name: "without escaping",
			giveEntry: DNSStaticEntry{
				Address: "127.0.0.1",
				Comment: "Any\\ text",
				Name:    "www.example\\.com",
				Regexp:  `.*\.example\.com`,
				TTL:     "1\\d",
			},
			wantString: `address=127.0.0.1 comment="Any\ text" disabled=no name="www.example\.com" regexp=".*\.example\.com" ttl="1\d"`, //nolint:lll
		},
		{
			name:       "empty",
			giveEntry:  DNSStaticEntry{},
			wantString: "",
			wantError:  ErrEmptyFields,
		},
		{
			name: "without address",
			giveEntry: DNSStaticEntry{
				Comment:  "Any text",
				Disabled: true,
				Name:     "www.example.com",
				Regexp:   `.*\.example\.com`,
				TTL:      "1d",
			},
			wantString: "",
			wantError:  ErrEmptyFields,
		},
		{
			name: "without hostname and regexp",
			giveEntry: DNSStaticEntry{
				Address:  "127.0.0.1",
				Comment:  "Any text",
				Disabled: true,
				TTL:      "1d",
			},
			wantString: "",
			wantError:  ErrEmptyFields,
		},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.giveEntry.Format(tt.givePrefix, tt.givePostfix)

			if tt.wantError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantError.Error())
			}

			assert.Equal(t, tt.wantString, string(res))
		})
	}
}
