package btrfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"

	"./fs"
	"./loop"
)

func TestIsSubvolume(t *testing.T) {
	b, err := IsSubvolume(".")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
}

func TestSubvolume(t *testing.T) {
	bf, err := backingFile()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(bf)
	lh, err := loop.AttachLoopDevice(bf)
	if err != nil {
		t.Fatal(err)
	}
	defer lh.Close()

	tmpdir, err := ioutil.TempDir("", "btrfs-testing.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	err = syscall.Mount(lh.Name(), tmpdir, "btrfs", uintptr(syscall.MS_NOATIME), "")
	if err != nil {
		t.Fatal(err)
	}
	defer syscall.Unmount(tmpdir, 0)

	for i := 0; i < 10; i++ {
		if err := SubvolCreate(tmpdir, fmt.Sprintf("vol-%d", i)); err != nil {
			t.Errorf("failed to create vol-%d : %s", i, err)
		}
	}
}

// helper
func backingFile() (string, error) {
	fh, err := ioutil.TempFile("", "btrfs-testing.")
	if err != nil {
		return "", err
	}

	// get a 1GB non-allocated file
	if err := fh.Truncate(1 * 1024 * 1024 * 1024); err != nil {
		return "", err
	}

	if err := fh.Close(); err != nil {
		return "", err
	}

	if err := fs.Mkfs(fh.Name()); err != nil {
		return "", err
	}
	return fh.Name(), nil
}
