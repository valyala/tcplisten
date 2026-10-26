//go:build linux || darwin || dragonfly || freebsd || netbsd || openbsd || rumprun
// +build linux darwin dragonfly freebsd netbsd openbsd rumprun

package tcplisten

import (
	"fmt"
	"net"
	"strconv"
	"syscall"
	"testing"
)

func TestGetSockaddrIPv6Zone(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	if len(interfaces) == 0 {
		t.Skip("no network interfaces")
	}
	iface := interfaces[0]
	for _, zone := range []string{iface.Name, strconv.Itoa(iface.Index)} {
		t.Run(zone, func(t *testing.T) {
			addr := fmt.Sprintf("[fe80::1%%%s]:8080", zone)
			sa, family, err := getSockaddr("tcp6", addr)
			if err != nil {
				t.Fatal(err)
			}
			sa6, ok := sa.(*syscall.SockaddrInet6)
			if !ok || family != syscall.AF_INET6 {
				t.Fatalf("unexpected sockaddr %T, family %d", sa, family)
			}
			if sa6.ZoneId != uint32(iface.Index) {
				t.Errorf("zone ID: got %d, want %d", sa6.ZoneId, iface.Index)
			}
			if !net.IP(sa6.Addr[:]).Equal(net.ParseIP("fe80::1")) || sa6.Port != 8080 {
				t.Errorf("unexpected address: %#v", sa6)
			}
		})
	}
}

func TestGetSockaddrInvalidIPv6Zone(t *testing.T) {
	for _, zone := range []string{"tcplisten-absent-interface", "-1", "4294967296"} {
		if _, _, err := getSockaddr("tcp6", "[fe80::1%"+zone+"]:8080"); err == nil {
			t.Errorf("accepted an invalid zone %q", zone)
		}
	}
}
