package hostsfile_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gh.tarampamp.am/mikrotik-hosts-parser/v4/pkg/hostsfile"
)

func BenchmarkParse(b *testing.B) {
	var benchDataset = []string{
		"../../test/testdata/hosts/ad_servers.txt",
		"../../test/testdata/hosts/foo.txt",
		"../../test/testdata/hosts/block_shit.txt",
		"../../test/testdata/hosts/hosts_adaway.txt",
		"../../test/testdata/hosts/serverlist.txt",
		"../../test/testdata/hosts/spy.txt",
	}

	for _, file := range benchDataset {
		for _, tc := range []struct {
			name       string
			isBuffered bool
		}{
			{"non buffered", false},
			{"buffered", true},
		} {
			name := strings.Join([]string{filepath.Base(file), tc.name}, " ")
			b.Run(name, func(b *testing.B) {
				fp, err := os.Open(file)
				if err != nil {
					b.Fatal(err)
				}
				defer fp.Close()

				raw, err := io.ReadAll(fp)
				if err != nil {
					b.Fatal(err)
				}

				buf := bytes.NewReader(raw)

				opts := make([]hostsfile.ParseOption, 0, 2)

				if tc.isBuffered {
					opts = append(
						opts,
						hostsfile.WithBufferSize(len(raw)),
						hostsfile.WithRecordsCount(len(raw)/36),
					)
				}

				b.SetBytes(int64(len(raw)))
				b.ReportAllocs()
				b.ResetTimer()

				for b.Loop() {
					b.StopTimer()
					buf.Seek(0, io.SeekStart)
					b.StartTimer()

					if _, e := hostsfile.Parse(buf, opts...); e != nil {
						b.Fatal(e)
					}
				}
			})
		}
	}
}

func TestParseUsingHostsFileContent(t *testing.T) {
	t.Parallel()

	var cases = []struct {
		giveFilePath  string
		wantRecords   int
		wantHostNames int
	}{
		{
			giveFilePath:  "../../test/testdata/hosts/ad_servers.txt",
			wantRecords:   45739,
			wantHostNames: 45739,
		},
		{
			giveFilePath:  "../../test/testdata/hosts/block_shit.txt",
			wantRecords:   109,
			wantHostNames: 109,
		},
		{
			giveFilePath:  "../../test/testdata/hosts/hosts_adaway.txt",
			wantRecords:   411,
			wantHostNames: 411,
		},
		{
			giveFilePath:  "../../test/testdata/hosts/hosts_malwaredomain.txt",
			wantRecords:   1106,
			wantHostNames: 1106,
		},
		{
			giveFilePath:  "../../test/testdata/hosts/hosts_someonewhocares.txt", // broken entry `127.0.0.1 secret.ɢoogle.com`
			wantRecords:   14308,
			wantHostNames: 14309, // ::1 [ip6-localhost ip6-loopback]
		},
		{
			giveFilePath:  "../../test/testdata/hosts/hosts_winhelp2002.txt",
			wantRecords:   11829,
			wantHostNames: 11829,
		},
		{
			giveFilePath:  "../../test/testdata/hosts/serverlist.txt",
			wantRecords:   3064,
			wantHostNames: 3064,
		},
		{
			giveFilePath:  "../../test/testdata/hosts/spy.txt",
			wantRecords:   367,
			wantHostNames: 367,
		},
	}

	for _, tt := range cases {
		tt := tt // reason: <https://git.io/fj8L6>
		t.Run("Hosts file: "+tt.giveFilePath, func(t *testing.T) {
			t.Parallel()

			file, err := os.Open(tt.giveFilePath)
			must(t, noError(t, err))
			defer file.Close()

			records, parseErr := hostsfile.Parse(file)
			must(t, noError(t, parseErr))

			must(t, equal(t, tt.wantRecords, len(records)))

			var hostsCount = 0

			for i := 0; i < len(records); i++ {
				if records[i].Host != "" {
					hostsCount++
				}

				hostsCount += len(records[i].AdditionalHosts)
			}

			equal(t, tt.wantHostNames, hostsCount)
		})
	}
}

func TestParseUsingCustomInput(t *testing.T) {
	buf := bytes.NewBuffer([]byte(`
# This is a sample
#comment without space
  #comment with spaces
	# comment with tab
1.2.3.4 dns.google #record with comment
4.3.2.1 bar.com # comment with space
4.3.2.1 ___id___.c.mystat-in.net		# comment with double tab
1.1.1.1 a.cn b.cn a.cn # "a.cn" is duplicate

::1  localfoo
2606:4700:4700::1111 cloudflare #[cf]

broken line format

0.0.0.1	example.com
0.0.0.1 example.com # duplicate

3.3.3.3	тест.рф xn--e1aybc.xn--p1ai

next line with IP only (spaces and tabs after)
0.0.0.0 		 		  	 	 		 	'

0 min.long.integer.ip               # valid
4294967295 max.long.integer.ip      # valid
4294967296 too-big.long.integer.ip  # invalid
-1 too-small.long.integer.ip        # invalid
4294N67295 broken.long.integer.ip   # invalid
# Space at the enc
127.0.0.1 ads.n-ws.org

the end
`))

	records, err := hostsfile.Parse(buf)
	must(t, equal(t, nil, err))

	var tt = [...]hostsfile.Record{
		{
			IP:   "1.2.3.4",
			Host: "dns.google",
		},
		{
			IP:   "4.3.2.1",
			Host: "bar.com",
		},
		{
			IP:   "4.3.2.1",
			Host: "___id___.c.mystat-in.net",
		},
		{
			IP:              "1.1.1.1",
			Host:            "a.cn",
			AdditionalHosts: []string{"b.cn", "a.cn"},
		},
		{
			IP:   "::1",
			Host: "localfoo",
		},
		{
			IP:   "2606:4700:4700::1111",
			Host: "cloudflare",
		},
		{
			IP:   "0.0.0.1",
			Host: "example.com",
		},
		{
			IP:   "0.0.0.1",
			Host: "example.com",
		},
		{ // "тест.рф" must be encoded as `xn--e1aybc.xn--p1ai`
			IP:   "3.3.3.3",
			Host: "xn--e1aybc.xn--p1ai",
		},
		{
			IP:   "0.0.0.0",
			Host: "min.long.integer.ip",
		},
		{
			IP:   "255.255.255.255",
			Host: "max.long.integer.ip",
		},
		{
			IP:   "127.0.0.1",
			Host: "ads.n-ws.org",
		},
	}

	must(t, equal(t, len(tt), len(records)))

	for i, want := range tt {
		equal(t, want, records[i])
	}
}

func equal(t *testing.T, want any, got any) bool {
	t.Helper()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("\nwant:\n%v\ngot:\n%v", want, got)

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

func must(t *testing.T, in bool) {
	if !in {
		t.FailNow()
	}
}
