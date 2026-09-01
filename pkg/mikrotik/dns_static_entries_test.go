package mikrotik

import (
	"bytes"
	"reflect"
	"testing"
)

//nolint:goconst // repeated fixture values keep rendering expectations explicit
func BenchmarkDNSStaticEntries_Render(b *testing.B) {
	b.ReportAllocs()

	data := make(DNSStaticEntries, 0, 1000)

	for range 1000 {
		data = append(data, DNSStaticEntry{
			Address:  "0.0.0.0",
			Comment:  "Any text",
			Disabled: true,
			Name:     "www.example.com",
			Regexp:   ".*\\.example\\.com",
			TTL:      "1d",
		})
	}

	var (
		i int
		e error
	)

	dest := bytes.NewBuffer([]byte{})

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		i, e = data.Render(dest)
	}

	if e != nil || i <= 0 || dest.Len() <= 0 {
		b.Fail()
	}
}

//nolint:goconst // repeated fixture values keep rendering expectations explicit
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
			wantResult: "foo address=0.0.0.0 disabled=no name=\"foo.com\" bar\nfoo address=8.8.8.8 disabled=no name=\"bar.com\" bar", //nolint:lll
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
			wantResult: `address=1.2.3.4 comment="foo comment" disabled=yes name="Bar name" regexp=".*\.example\.com" ttl="1d"`, //nolint:lll
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

			equal(t, len(tt.wantResult), l)

			if tt.wantError != nil {
				if err == nil {
					t.Errorf("got error %v, want error nil", tt.wantError)
				} else {
					equal(t, err.Error(), tt.wantError.Error())
				}
			} else {
				noError(t, err)
			}

			equal(t, tt.wantResult, buf.String())
		})
	}
}

func equal(t *testing.T, got, want any) bool {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
		return false
	}

	return true
}

func noError(t *testing.T, err error) bool {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error %v", err)
		return false
	}

	return true
}
