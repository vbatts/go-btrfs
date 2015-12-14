package btrfs

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
