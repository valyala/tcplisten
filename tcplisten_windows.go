// +build windows

package tcplisten

import (
	"net"
)

// A dummy implementation for windows
type Config struct {
	ReusePort   bool
	DeferAccept bool
	FastOpen    bool
	Backlog     int
}

func (cfg *Config) NewListener(network, addr string) (net.Listener, error) {
	return net.Listen(network, addr)
}
