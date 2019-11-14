package parser

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func BenchmarkParser_ParseLargeFile(b *testing.B) {
	file, err := os.Open(".tests/hosts/ad_servers.txt")
	if err != nil {
		panic(err)
	}

	for n := 0; n < b.N; n++ {
		_, _ = (&Parser{}).Parse(file)
	}

	if err := file.Close(); err != nil {
		panic(err)
	}
}

func TestHostsSourceParser_ParseHostsFileUsingTestData(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		filepath      string
		wantRecords   int
		wantHostNames int
	}{
		{
			filepath:      "../../.tests/hosts/ad_servers.txt",
			wantRecords:   45739,
			wantHostNames: 45739,
		},
		{
			filepath:      "../../.tests/hosts/block_shit.txt",
			wantRecords:   109,
			wantHostNames: 109,
		},
		{
			filepath:      "../../.tests/hosts/hosts_adaway.txt",
			wantRecords:   411,
			wantHostNames: 411,
		},
		{
			filepath:      "../../.tests/hosts/hosts_malwaredomain.txt",
			wantRecords:   1106,
			wantHostNames: 1106,
		},
		{
			filepath:      "../../.tests/hosts/hosts_someonewhocares.txt", // broken entry `127.0.0.1 secret.ɢoogle.com`
			wantRecords:   14308,
			wantHostNames: 14309, // ::1 [ip6-localhost ip6-loopback]
		},
		{
			filepath:      "../../.tests/hosts/hosts_winhelp2002.txt",
			wantRecords:   11829,
			wantHostNames: 11829,
		},
		{
			filepath:      "../../.tests/hosts/serverlist.txt",
			wantRecords:   3064,
			wantHostNames: 3064,
		},
		{
			filepath:      "../../.tests/hosts/spy.txt",
			wantRecords:   367,
			wantHostNames: 367,
		},
	}

	for _, tt := range cases {
		t.Run("Using file "+tt.filepath, func(t *testing.T) {
			t.Parallel()

			file, err := os.Open(tt.filepath)
			if err != nil {
				panic(err)
			}

			res, parseErr := (&Parser{}).Parse(file)

			if parseErr != nil {
				t.Error(parseErr)
			}

			if rLen := len(res); rLen != tt.wantRecords {
				t.Errorf("Expected records count is %d, got %d", tt.wantRecords, rLen)
			}

			var hostsCount int = 0
			for _, p := range res {
				hostsCount += len(p.Hosts)
				if len(p.Hosts) > 1 {
					fmt.Println(p.IP, p.Hosts)
				}
			}

			if hostsCount != tt.wantHostNames {
				t.Errorf("Expected hosts count is %d, got %d", tt.wantHostNames, hostsCount)
			}

			if err := file.Close(); err != nil {
				panic(err)
			}
		})
	}
}

func TestParser_validateHostname(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		hostname   string
		wantResult bool
	}{
		{
			hostname:   "www.google.com",
			wantResult: true,
		},
		{
			hostname:   "google.com",
			wantResult: true,
		},
		{
			hostname:   "dns.google",
			wantResult: true,
		},
		{
			hostname:   "localhost",
			wantResult: true,
		},
		{
			hostname:   "x",
			wantResult: true,
		},
		{
			hostname:   "i.oh1.me",
			wantResult: true,
		},
		{
			hostname:   "localhost",
			wantResult: true,
		},
		{
			hostname:   "ip6-loopback",
			wantResult: true,
		},
		{
			hostname:   "ad-g.doubleclick.net",
			wantResult: true,
		},
		{
			hostname:   "sO.2mdn.net",
			wantResult: true,
		},
		{
			hostname:   "adman_test.go2_cloud.org",
			wantResult: true,
		},
		{
			hostname:   "r1---sn-vgqsen7z.googlevideo.com",
			wantResult: true,
		},
		{
			hostname:   "xn--90a5ai.xn--p1ai",
			wantResult: true,
		},
		{
			hostname:   "foo.bar.baz.123.com",
			wantResult: true,
		},
		{
			hostname:   "___id___.c.mystat-in.net",
			wantResult: true,
		},
		{
			hostname:   "goo gle.com",
			wantResult: false,
		},
		{
			hostname:   "\\.com",
			wantResult: false,
		},
		{
			hostname:   "/.com",
			wantResult: false,
		},
		{
			hostname:   "тест.рф", // must be encoded in `xn--e1aybc.xn--p1ai`
			wantResult: false,
		},
	}

	parser := &Parser{}

	for _, tt := range cases {
		t.Run("Using "+tt.hostname, func(t *testing.T) {
			if res := parser.validateHostname(tt.hostname); res != tt.wantResult {
				t.Errorf(
					`For "%s" must returns "%v", but returns "%v"`,
					tt.hostname,
					tt.wantResult,
					res,
				)
			}
		})
	}
}

func TestParser_parseRawLine(t *testing.T) { //nolint:funlen
	t.Parallel()

	var cases = []struct {
		line       string
		wantError  error
		wantResult bool
		wantIP     string
		wantHosts  []string
	}{
		{
			line:       "127.0.0.1 google.com dns.google",
			wantResult: true,
			wantIP:     "127.0.0.1",
			wantHosts:  []string{"google.com", "dns.google"},
		},
		{
			line:       "0.0.0.0 ___id___.c.mystat-in.net",
			wantResult: true,
			wantIP:     "0.0.0.0",
			wantHosts:  []string{"___id___.c.mystat-in.net"},
		},
		{
			line:       "   127.0.0.1 \t\tgoogle.com\tdns.google  ",
			wantResult: true,
			wantIP:     "127.0.0.1",
			wantHosts:  []string{"google.com", "dns.google"},
		},
		{
			line:       "   fe80::74e6:b5f3:fe92:830e \t\tgoogle.com\tdns.google  ",
			wantResult: true,
			wantIP:     "fe80::74e6:b5f3:fe92:830e",
			wantHosts:  []string{"google.com", "dns.google"},
		},
		{
			line:       "::1 google.com dns.google  ",
			wantResult: true,
			wantIP:     "::1",
			wantHosts:  []string{"google.com", "dns.google"},
		},
		{
			line:       "foo",
			wantError:  errors.New("hosts line parser: wrong line format"),
			wantResult: false,
		},
		{
			line:       "foo bar",
			wantError:  errors.New("hosts line parser: wrong IP address"),
			wantResult: false,
		},
		{
			line:       "8.8.8.8 ^",
			wantError:  errors.New("hosts line parser: hosts not found"),
			wantResult: false,
		},
		{
			line:       "   1.1.1.257 bar",
			wantError:  errors.New("hosts line parser: wrong IP address"),
			wantResult: false,
		},
		{
			line:       "#127.0.0.1 google.com dns.google",
			wantResult: false,
		},
		{
			line:       "  #127.0.0.1 google.com dns.google",
			wantResult: false,
		},
		{
			line:       "#",
			wantResult: false,
		},
		{
			line:       " #",
			wantResult: false,
		},
		{
			line:       "# ",
			wantResult: false,
		},
		{
			line:       "# Comment line",
			wantResult: false,
		},
		{
			line:       "### Comment line",
			wantResult: false,
		},
		{
			line:       " ### Comment line",
			wantResult: false,
		},
		{
			line:       "127.0.0.1 google.com dns.google # some comment",
			wantResult: true,
			wantIP:     "127.0.0.1",
			wantHosts:  []string{"google.com", "dns.google"},
		},
		{
			line:       "127.0.0.1 google.com dns.google #some comment localhost",
			wantResult: true,
			wantIP:     "127.0.0.1",
			wantHosts:  []string{"google.com", "dns.google"},
		},
		{
			line:       "0.0.0.0 xn--90a5ai.xn--p1ai\tx \\.com",
			wantResult: true,
			wantIP:     "0.0.0.0",
			wantHosts:  []string{"xn--90a5ai.xn--p1ai", "x"},
		},
		{
			line:       "2001:db8:0:1:1:1:1:1 xn--90a5ai.xn--p1ai\tx \\.com",
			wantResult: true,
			wantIP:     "2001:db8:0:1:1:1:1:1",
			wantHosts:  []string{"xn--90a5ai.xn--p1ai", "x"},
		},
	}

	for _, tt := range cases {
		t.Run("Using "+tt.line, func(t *testing.T) {
			res, err := (&Parser{}).parseRawLine(tt.line)

			if tt.wantError != nil && err.Error() != tt.wantError.Error() {
				t.Errorf(`Want error "%v", but got "%v"`, tt.wantError, err)
			}

			if err != nil && tt.wantError == nil {
				t.Errorf(`Error %v returned, but nothing expected`, err)
			}

			if tt.wantResult && res != nil {
				if tt.wantIP != res.IP.String() {
					t.Errorf(`Want IP "%s", but got "%s"`, tt.wantIP, res.IP)
				}

				if !reflect.DeepEqual(tt.wantHosts, res.Hosts) {
					t.Errorf("Want hosts %v, but got %v", tt.wantHosts, res.Hosts)
				}
			}

			if tt.wantResult && res == nil {
				t.Error("Expected non-nil result, but nil")
			}
		})
	}
}

func TestParser_startsWithRune(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		giveString string
		giveRune   rune
		wantResult bool
	}{
		{
			giveString: "! foo",
			giveRune:   '!',
			wantResult: true,
		},
		{
			giveString: " ! foo",
			giveRune:   '!',
			wantResult: false,
		},
		{
			giveString: "abracadabra",
			giveRune:   'a',
			wantResult: true,
		},
		{
			giveString: "",
			giveRune:   'a',
			wantResult: false,
		},
		{
			giveString: "",
			giveRune:   ' ',
			wantResult: false,
		},
	}

	for _, tt := range cases {
		t.Run("Using "+tt.giveString, func(t *testing.T) {
			res := (&Parser{}).startsWithRune(tt.giveString, tt.giveRune)

			if tt.wantResult != res {
				t.Errorf(`Want result "%v", but got "%v"`, tt.wantResult, res)
			}
		})
	}
}
