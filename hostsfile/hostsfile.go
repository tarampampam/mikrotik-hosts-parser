package hostsfile

import (
	"net"
)

// Hosts file record
type Record struct {
	IP    net.IP
	Hosts []string
}
