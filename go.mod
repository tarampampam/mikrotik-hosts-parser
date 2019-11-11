module mikrotik-hosts-parser

go 1.13

require (
	github.com/gorilla/mux v1.7.3
	github.com/jessevdk/go-flags v1.4.0
)

replace (
	github.com/tarampampam/mikrotik-hosts-parser/hostsfile => ../hostsfile
	github.com/tarampampam/mikrotik-hosts-parser/hostsfile/parser => ../hostsfile/parser
	github.com/tarampampam/mikrotik-hosts-parser/http/fileserver => ../http/fileserver
	github.com/tarampampam/mikrotik-hosts-parser/mikrotik/dns => ../mikrotik/dns
	github.com/tarampampam/mikrotik-hosts-parser/options => ../options
	github.com/tarampampam/mikrotik-hosts-parser/resources => ../resources
)
