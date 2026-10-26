//go:build linux
// +build linux

package tcplisten

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReadSoMaxConnPermissionDenied(t *testing.T) {
	file, err := ioutil.TempFile("", "tcplisten-somaxconn-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(file.Name(), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := ioutil.ReadFile(file.Name()); err == nil {
		t.Skip("current user can read mode-000 files")
	} else if !os.IsPermission(err) {
		t.Fatal(err)
	}
	backlog, err := readSoMaxConn(file.Name())
	if err != nil {
		t.Fatal(err)
	}
	if backlog != syscall.SOMAXCONN {
		t.Errorf("backlog: got %d, want %d", backlog, syscall.SOMAXCONN)
	}
}

func TestReadSoMaxConn(t *testing.T) {
	dir, err := ioutil.TempDir("", "tcplisten-somaxconn-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "somaxconn")
	backlog, err := readSoMaxConn(path)
	if err != nil || backlog != syscall.SOMAXCONN {
		t.Fatalf("missing sysctl: backlog %d, err %v", backlog, err)
	}
	if err := ioutil.WriteFile(path, []byte("1234\n"), 0600); err != nil {
		t.Fatal(err)
	}
	backlog, err = readSoMaxConn(path)
	if err != nil || backlog != 1234 {
		t.Fatalf("readable sysctl: backlog %d, err %v", backlog, err)
	}
	if err := ioutil.WriteFile(path, []byte("invalid"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := readSoMaxConn(path); err == nil {
		t.Error("accepted malformed sysctl")
	}
	if _, err := readSoMaxConn(dir); err == nil {
		t.Error("accepted directory as sysctl")
	}
}
