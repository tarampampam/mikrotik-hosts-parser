package main

type HostsSourceParser struct {
	RedirectTo    string
	ExcludedHosts []string
}

type MikrotikDnsStaticEntry struct {
	Address  string `comment:"IP address" property:"address" examples:"0.0.0.0"`
	Comment  string `comment:"Short description of the item" property:"comment" examples:"Any text"`
	Disabled bool   `comment:"Defines whether item is ignored or used" property:"disabled" examples:"yes,no"`
	Name     string `comment:"Host name" property:"name" examples:"www.example.com"`
	Regexp   string `property:"regexp" examples:".*\\.example\\.com"`
	TTL      string `comment:"Time To Live" property:"ttl" examples:"1d"` // @todo: Need more examples
}
