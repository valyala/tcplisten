// +build linux

package tcplisten

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

const (
	soReusePort = 0x0F
	tcpFastOpen = 0x17
)

func enableDeferAccept(fd int) error {
	if err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_DEFER_ACCEPT, 1); err != nil {
		return fmt.Errorf("cannot enable TCP_DEFER_ACCEPT: %s", err)
	}
	return nil
}

func enableFastOpen(fd int) error {
	if err := syscall.SetsockoptInt(fd, syscall.SOL_TCP, tcpFastOpen, fastOpenQlen); err != nil {
		return fmt.Errorf("cannot enable TCP_FASTOPEN(qlen=%d): %s", fastOpenQlen, err)
	}
	return nil
}

const fastOpenQlen = 16 * 1024

func soMaxConn() (int, error) {
	data, err := ioutil.ReadFile(soMaxConnFilePath)
	if err != nil {
		// This error may trigger on travis build. Just use SOMAXCONN
		if os.IsNotExist(err) {
			return syscall.SOMAXCONN, nil
		}
		return -1, err
	}
	s := strings.TrimSpace(string(data))
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return -1, fmt.Errorf("cannot parse somaxconn %q read from %s: %s", s, soMaxConnFilePath, err)
	}

	if n > 1<<16-1 {
		return maxAckBacklog(n), nil
	}
	return n, nil
}

// Linux stores the backlog as:
//
//   - uint16 in kernel version < 4.1,
//   - uint32 in kernel version >= 4.1
//
// Truncate number to avoid wrapping.
//
// See https://github.com/golang/go/issues/41470 and https://github.com/golang/go/issues/5030.
func maxAckBacklog(n int) int {
	major, minor := kernelVersion()
	return maxAckBacklogForVersion(n, major, minor)
}

func maxAckBacklogForVersion(n int, major, minor int) int {
	var max uint = 1<<16 - 1
	if major > 4 || (major == 4 && minor >= 1) {
		max = 1<<32 - 1
	}

	if uint(n) > max {
		n = int(max)
	}
	return n
}

func kernelVersion() (major, minor int) {
	var uname syscall.Utsname
	if err := syscall.Uname(&uname); err != nil {
		return 0, 0
	}
	var buf [65]byte
	var n int
	for _, c := range uname.Release {
		if c == 0 {
			break
		}
		buf[n] = byte(c)
		n++
	}
	return parseKernelVersion(string(buf[:n]))
}

func parseKernelVersion(release string) (major, minor int) {
	var (
		values    [2]int
		value, vi int
	)
	for _, c := range release {
		if '0' <= c && c <= '9' {
			value = (value * 10) + int(c-'0')
		} else {
			values[vi] = value
			vi++
			if vi >= len(values) {
				break
			}
			value = 0
		}
	}
	if vi < len(values) {
		values[vi] = value
	}
	return values[0], values[1]
}

var soMaxConnFilePath = "/proc/sys/net/core/somaxconn"
