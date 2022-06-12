package hostsfile

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var benchDataset = []struct{ filePath string }{ //nolint:gochecknoglobals
	{"../../test/testdata/hosts/foo.txt"},
	{"../../test/testdata/hosts/ad_servers.txt"},
	{"../../test/testdata/hosts/block_shit.txt"},
	{"../../test/testdata/hosts/hosts_adaway.txt"},
	{"../../test/testdata/hosts/serverlist.txt"},
	{"../../test/testdata/hosts/spy.txt"},
}

func BenchmarkParse(b *testing.B) {
	for _, tt := range benchDataset {
		tt := tt

		b.Run(filepath.Base(tt.filePath), func(b *testing.B) {
			b.ReportAllocs()

			raw, err := ioutil.ReadFile(tt.filePath)
			if err != nil {
				panic(err)
			}

			b.SetBytes(int64(len(raw)))
			b.ResetTimer()
			for n := 0; n < b.N; n++ {
				b.StopTimer()
				buf := bytes.NewBuffer(raw)
				b.StartTimer()

				_, e := Parse(buf)

				if e != nil {
					b.Fatal(e)
				}
			}
		})
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
			assert.NoError(t, err)

			records, parseErr := Parse(file)
			assert.NoError(t, file.Close())
			assert.NoError(t, parseErr)

			assert.Len(t, records, tt.wantRecords)

			var hostsCount = 0

			for i := 0; i < len(records); i++ {
				if records[i].Host != "" {
					hostsCount++
				}

				hostsCount += len(records[i].AdditionalHosts)
			}

			assert.Equal(t, tt.wantHostNames, hostsCount)
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

the end
`))

	records, err := Parse(buf)
	assert.NoError(t, err)

	assert.Len(t, records, 11)

	assert.Equal(t, "1.2.3.4", records[0].IP)
	assert.Equal(t, "dns.google", records[0].Host)
	assert.Nil(t, records[0].AdditionalHosts)

	assert.Equal(t, "4.3.2.1", records[1].IP)
	assert.Equal(t, "bar.com", records[1].Host)
	assert.Nil(t, records[1].AdditionalHosts)

	assert.Equal(t, "4.3.2.1", records[2].IP)
	assert.Equal(t, "___id___.c.mystat-in.net", records[2].Host)
	assert.Nil(t, records[2].AdditionalHosts)

	assert.Equal(t, "1.1.1.1", records[3].IP)
	assert.Equal(t, "a.cn", records[3].Host)
	assert.ElementsMatch(t, []string{"b.cn", "a.cn"}, records[3].AdditionalHosts)

	assert.Equal(t, "::1", records[4].IP)
	assert.Equal(t, "localfoo", records[4].Host)
	assert.Nil(t, records[4].AdditionalHosts)

	assert.Equal(t, "2606:4700:4700::1111", records[5].IP)
	assert.Equal(t, "cloudflare", records[5].Host)
	assert.Nil(t, records[5].AdditionalHosts)

	assert.Equal(t, "0.0.0.1", records[6].IP)
	assert.Equal(t, "example.com", records[6].Host)
	assert.Nil(t, records[6].AdditionalHosts)

	assert.Equal(t, "0.0.0.1", records[7].IP)
	assert.Equal(t, "example.com", records[7].Host)
	assert.Nil(t, records[7].AdditionalHosts)

	// "тест.рф" must be encoded as `xn--e1aybc.xn--p1ai`
	assert.Equal(t, "3.3.3.3", records[8].IP)
	assert.Equal(t, "xn--e1aybc.xn--p1ai", records[8].Host)
	assert.Nil(t, records[8].AdditionalHosts)

	assert.Equal(t, "0.0.0.0", records[9].IP) // long: 0
	assert.Equal(t, "min.long.integer.ip", records[9].Host)
	assert.Nil(t, records[7].AdditionalHosts)

	assert.Equal(t, "255.255.255.255", records[10].IP) // long: 4294967295
	assert.Equal(t, "max.long.integer.ip", records[10].Host)
	assert.Nil(t, records[7].AdditionalHosts)
}
