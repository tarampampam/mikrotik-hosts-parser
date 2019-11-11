package hostsfile

import (
	"net"
	"reflect"
	"testing"
)

func TestRecord(t *testing.T) {
	t.Parallel()

	r := Record{
		IP:    net.IPv4(127, 0, 0, 1),
		Hosts: []string{"localhost"},
	}

	if r.IP.String() != "127.0.0.1" {
		t.Errorf("Wrong IP addr: %v", r.IP)
	}

	if !reflect.DeepEqual(r.Hosts, []string{"localhost"}) {
		t.Errorf("Wrong hosts: %v", r.Hosts)
	}
}
