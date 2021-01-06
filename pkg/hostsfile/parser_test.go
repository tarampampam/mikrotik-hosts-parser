package hostsfile

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkParse(b *testing.B) { // 1851661	       567 ns/op	    4106 B/op	       1 allocs/op
	b.ReportAllocs()

	raw, err := ioutil.ReadFile("../../test/testdata/hosts/ad_servers.txt")
	if err != nil {
		panic(err)
	}

	buf := bytes.NewBuffer(raw)

	for n := 0; n < b.N; n++ {
		_, _ = Parse(buf)
	}
}

func TestParseUsingHostsFileContent(t *testing.T) {
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
				hostsCount += len(records[i].Hosts)
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

the end
`))

	records, err := Parse(buf)

	assert.Len(t, records, 9)

	assert.ElementsMatch(t, []string{"dns.google"}, records[0].Hosts)
	assert.ElementsMatch(t, []string{"bar.com"}, records[1].Hosts)
	assert.ElementsMatch(t, []string{"___id___.c.mystat-in.net"}, records[2].Hosts)
	assert.ElementsMatch(t, []string{"a.cn", "b.cn", "a.cn"}, records[3].Hosts)
	assert.ElementsMatch(t, []string{"localfoo"}, records[4].Hosts)
	assert.ElementsMatch(t, []string{"cloudflare"}, records[5].Hosts)
	assert.ElementsMatch(t, []string{"example.com"}, records[6].Hosts)
	assert.ElementsMatch(t, []string{"example.com"}, records[7].Hosts)
	assert.ElementsMatch(t, []string{"xn--e1aybc.xn--p1ai"}, records[8].Hosts) // "тест.рф" must be encoded in `xn--e1aybc.xn--p1ai`

	t.Log(records, err)
}
