package btrfs

import (
	"io/ioutil"
	"testing"

	"./fs"
)

func TestIsSubvolume(t *testing.T) {
	b, err := IsSubvolume(".")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(b)
}

func testingImage() (string, error) {
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
