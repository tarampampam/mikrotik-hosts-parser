package hostsfile

import "net"

// Record is a hosts file record.
type Record struct {
	IP    net.IP
	Hosts []string
}
