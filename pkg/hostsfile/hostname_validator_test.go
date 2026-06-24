package hostsfile

import (
	"strings"
	"testing"
)

func TestValidateDomainZeroAlloc_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		isInvalid bool
	}{
		{"Valid standard ASCII", "example.com", false},
		{"Valid single letter label", "a.b.c.de", false},
		{"Valid Punycode", "xn--mase-qka.com", false},
		{"Valid complex Punycode", "xn--80akhbyknj4f.xn--p1ai", false},
		{"Valid numbers and hyphens", "my-short-123-domain.net", false},
		{
			"Valid maximum domain length",
			strings.Repeat("a", 63) +
				"." + strings.Repeat("a", 63) +
				"." + strings.Repeat("a", 63) +
				"." + strings.Repeat("a", 60),
			false,
		},

		{"Empty string", "", true},
		{"Label too long (>63 chars)", "a." + strings.Repeat("x", 64) + ".com", true},
		{
			"Domain too long (>253 chars)",
			strings.Repeat("a", 63) +
				"." + strings.Repeat("a", 63) +
				"." + strings.Repeat("a", 63) +
				"." + strings.Repeat("a", 62),
			true,
		},
		{"Empty label at start", ".example.com", true},
		{"Empty label in middle", "example..com", true},
		{"Empty label at end", "example.com.", true},

		{"Label starts with hyphen", "-example.com", true},
		{"Label ends with hyphen", "example-.com", true},
		{"Subdomain starts with hyphen", "://-example.com", true},
		{"Subdomain ends with hyphen", "://example-.com", true},
		{"Only hyphens in label", "---.com", true},

		{"Punycode prefix with zero payload", "xn--.com", true},
		{"Punycode looks valid but ends in hyphen (CVE target)", "xn--example-.com", true},
		{"Punycode inner label ends in hyphen", "xn--mase-qka-.com", true},
		{"Punycode with nested invalid double hyphen", "xn--ab--cd.com", true},
		{"Punycode uppercase variations (Should be structural valid)", "XN--mase-qka.com", false},
		{"Punycode mixed case prefix", "xN--mase-qka.com", false},
		{"Punycode invalid character in payload", "xn--mas_e-qka.com", true},
		{"Punycode lookalike but short", "xn--a.com", false},

		{"Spaces inside domain", "ex ample.com", true},
		{"Leading space", " example.com", true},
		{"Trailing space", "example.com ", true},
		{"Exclamation mark", "example!.com", true},
		{"Null byte injection attempt", "example\x00.com", true},
		// 'е' Cyrillic small letter ie (U+0435) vs Latin small letter e (U+0065)
		{"Unicode lookalike homograph raw", "еxample.com", true},
		{"Path injection attempt", "://example.com", true},
		{"Port injection attempt", "example.com:8080", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if validateHostname([]byte(tt.domain)) == tt.isInvalid {
				var s string
				if tt.isInvalid {
					s = "invalid"
				} else {
					s = "valid"
				}
				t.Errorf("domain [%s] should be %s", tt.domain, s)
			}
		})
	}
}

func BenchmarkValidateDomainZeroAlloc(b *testing.B) {
	domain := []byte("my-very-long-edge-case-domain-specification.com")

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		if !validateHostname(domain) {
			b.Fatal("returned false")
		}
	}
}
