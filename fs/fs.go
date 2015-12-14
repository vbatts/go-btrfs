package fs

import (
	"os"
	"os/exec"
)

var (
	// DefaultMkfsBtrfs is the default binary used to Mkfs
	DefaultMkfsBtrfs = "/usr/sbin/mkfs.btrfs"
)

// Mkfs will initialize a btrfs filesystem on file or block device at filepath (with only default options)
func Mkfs(filepath ...string) error {
	cmd := exec.Command(DefaultMkfsBtrfs, filepath...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

// TruncateBtrfs is a convenience method to grow filename to size bytes
func TruncateBtrfs(filename string, size int64) error {
	fh, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	if err := fh.Truncate(size); err != nil {
		return err
	}
	return fh.Close()
}
