// +build linux

package tcplisten

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestParseKernelVersion(t *testing.T) {
	testCases := []struct {
		release string
		major   int
		minor   int
	}{
		{"5.4.0-42-generic", 5, 4},
		{"4.1.0", 4, 1},
		{"4.1", 4, 1},
		{"4.0.0", 4, 0},
		{"4", 4, 0},
		{"3.10.0-1160.el7.x86_64", 3, 10},
		{"6.1.0-28-amd64", 6, 1},
		{"7.0.14-generic", 7, 0},
		{"", 0, 0},
		{"invalid", 0, 0},
	}

	for _, tc := range testCases {
		major, minor := parseKernelVersion(tc.release)
		if major != tc.major || minor != tc.minor {
			t.Errorf("parseKernelVersion(%q) = (%d, %d), want (%d, %d)", tc.release, major, minor, tc.major, tc.minor)
		}
	}
}

func TestMaxAckBacklogForVersion(t *testing.T) {
	testCases := []struct {
		name     string
		n        int
		major    int
		minor    int
		expected int
	}{
		{"kernel 3.10 truncates to uint16", 196602, 3, 10, 1<<16 - 1},
		{"kernel 4.0 truncates to uint16", 196602, 4, 0, 1<<16 - 1},
		{"kernel 4.1 keeps uint32 size", 196602, 4, 1, 196602},
		{"kernel 5.4 keeps uint32 size", 196602, 5, 4, 196602},
		{"kernel 6.1 keeps uint32 size", 196602, 6, 1, 196602},
		{"kernel 7.0 keeps uint32 size", 196602, 7, 0, 196602},
		{"small value preserved on kernel 3.10", 128, 3, 10, 128},
		{"small value preserved on kernel 5.4", 128, 5, 4, 128},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := maxAckBacklogForVersion(tc.n, tc.major, tc.minor)
			if got != tc.expected {
				t.Errorf("maxAckBacklogForVersion(%d, %d, %d) = %d, want %d", tc.n, tc.major, tc.minor, got, tc.expected)
			}
		})
	}
}

func TestMaxAckBacklog(t *testing.T) {
	n := 196602
	backlog := maxAckBacklog(n)
	major, minor := kernelVersion()
	expected := 1<<16 - 1
	if major > 4 || (major == 4 && minor >= 1) {
		expected = n
	}
	if backlog != expected {
		t.Fatalf("unexpected backlog: got %d, want %d", backlog, expected)
	}

	if got := maxAckBacklog(128); got != 128 {
		t.Fatalf("unexpected backlog for small value: got %d, want 128", got)
	}
}

func TestSoMaxConnWithLargeValue(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "somaxconn_test")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("196602\n"); err != nil {
		t.Fatalf("cannot write to temp file: %s", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("cannot close temp file: %s", err)
	}

	oldPath := soMaxConnFilePath
	soMaxConnFilePath = tmpFile.Name()
	defer func() {
		soMaxConnFilePath = oldPath
	}()

	n, err := soMaxConn()
	if err != nil {
		t.Fatalf("unexpected error from soMaxConn: %s", err)
	}

	major, minor := kernelVersion()
	expected := 1<<16 - 1
	if major > 4 || (major == 4 && minor >= 1) {
		expected = 196602
	}
	if n != expected {
		t.Fatalf("soMaxConn() = %d, want %d", n, expected)
	}
}
