//go:build darwin || dragonfly || freebsd || netbsd || openbsd || rumprun
// +build darwin dragonfly freebsd netbsd openbsd rumprun

package tcplisten

import (
	"fmt"
	"syscall"
	"time"
)

const soReusePort = syscall.SO_REUSEPORT

func enableKeepAlive(fd int, d time.Duration) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_KEEPALIVE, 1); err != nil {
		return fmt.Errorf("cannot enable SO_KEEPALIVE: %s", err)
	}
	return nil
}

func enableDeferAccept(fd int) error {
	// TODO: implement SO_ACCEPTFILTER:dataready here
	return nil
}

func enableFastOpen(fd int) error {
	// TODO: implement TCP_FASTOPEN when it will be ready
	return nil
}

func soMaxConn() (int, error) {
	// TODO: properly implement it
	return syscall.SOMAXCONN, nil
}
