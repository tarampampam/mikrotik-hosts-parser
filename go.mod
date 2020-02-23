module mikrotik-hosts-parser

go 1.13

require (
	github.com/a8m/envsubst v1.1.0
	github.com/gorilla/mux v1.7.4
	github.com/jessevdk/go-flags v1.4.1-0.20181221193153-c0795c8afcf4
	github.com/kr/pretty v0.2.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/tarampampam/go-filecache v1.0.2
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8
)

replace (
	github.com/tarampampam/mikrotik-hosts-parser/cmd => ../cmd
	github.com/tarampampam/mikrotik-hosts-parser/cmd/serve => ../cmd/serve
	github.com/tarampampam/mikrotik-hosts-parser/cmd/version => ../cmd/version
	github.com/tarampampam/mikrotik-hosts-parser/hostsfile => ../hostsfile
	github.com/tarampampam/mikrotik-hosts-parser/hostsfile/parser => ../hostsfile/parser
	github.com/tarampampam/mikrotik-hosts-parser/http => ../http
	github.com/tarampampam/mikrotik-hosts-parser/http/api => ../http/api
	github.com/tarampampam/mikrotik-hosts-parser/http/fileserver => ../http/fileserver
	github.com/tarampampam/mikrotik-hosts-parser/http/script => ../http/script
	github.com/tarampampam/mikrotik-hosts-parser/mikrotik/dns => ../mikrotik/dns
	github.com/tarampampam/mikrotik-hosts-parser/settings/serve => ../settings/serve
	github.com/tarampampam/mikrotik-hosts-parser/version => ../version
)
