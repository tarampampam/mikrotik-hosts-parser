module mikrotik-hosts-parser

go 1.13

require (
	github.com/a8m/envsubst v1.1.0
	github.com/gorilla/mux v1.7.3
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
	gopkg.in/yaml.v2 v2.2.5
)

replace (
	github.com/tarampampam/mikrotik-hosts-parser/cmd => ../cmd
	github.com/tarampampam/mikrotik-hosts-parser/cmd/serve => ../cmd/serve
	github.com/tarampampam/mikrotik-hosts-parser/cmd/version => ../cmd/version
	github.com/tarampampam/mikrotik-hosts-parser/hostsfile => ../hostsfile
	github.com/tarampampam/mikrotik-hosts-parser/hostsfile/parser => ../hostsfile/parser
	github.com/tarampampam/mikrotik-hosts-parser/http => ../http
	github.com/tarampampam/mikrotik-hosts-parser/http/fileserver => ../http/fileserver
	github.com/tarampampam/mikrotik-hosts-parser/mikrotik/dns => ../mikrotik/dns
	github.com/tarampampam/mikrotik-hosts-parser/options => ../options
	github.com/tarampampam/mikrotik-hosts-parser/resources => ../resources
	github.com/tarampampam/mikrotik-hosts-parser/settings => ../settings
)
