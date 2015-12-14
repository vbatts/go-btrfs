package btrfs

import (
	"io/ioutil"
	"os"
	"testing"

	"./loop"
)

func TestIsSubvolume(t *testing.T) {
	b, err := IsSubvolume(".")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%q", b)
}

func TestMkfs(t *testing.T) {
	fh, err := ioutil.TempFile("", "btrfs-testing.")
	if err != nil {
		t.Error(err)
	}

	// get a 1GB non-allocated file
	if err := fh.Truncate(1 * 1024 * 1024 * 1024); err != nil {
		t.Error(err)
	}

	if err := fh.Close(); err != nil {
		t.Error(err)
	}
	defer os.Remove(fh.Name())

	if err := Mkfs(fh.Name()); err != nil {
		t.Error(err)
	}
	_ = loop.AttachLoopDevice
}
