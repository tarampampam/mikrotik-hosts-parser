package hostsfile

// Record is a hosts file record.
type Record struct {
	IP              string
	Host            string
	AdditionalHosts []string
}
