// +build linux

/*
Initial source imported from github.com/docker/docker ./daemon/graphdriver/btrfs/ @ 75d69ce0da2e360773736502acd92d4a9cf7faa5
See LICENSE.docker
*/

package btrfs

/*
#include <stdlib.h>
#include <btrfs/ioctl.h>
#include <btrfs/ctree.h>
*/
import "C"

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	// ErrorNameTooLong is returned if the name is longer than the default 255 characters
	ErrorNameTooLong = errors.New("name length too long")
)

// SubvolCreate creates a new btrfs subvolume, with dirpath being the root directory
func SubvolCreate(dirpath, name string) error {
	dir, err := openDir(dirpath)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	if len(name) > C.BTRFS_NAME_LEN {
		return ErrorNameTooLong
	}
	var args C.struct_btrfs_ioctl_vol_args
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_SUBVOL_CREATE,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to create btrfs subvolume: %v", errno.Error())
	}
	return nil
}

// SubvolSnapshot creates a new btrfs subvolume snapshot. With `dirpath` being
// the root directory, make a snapshot of `srcdirpath`, with name `name`.
func SubvolSnapshot(srcdirpath, dirpath, name string) error {
	srcDir, err := openDir(srcdirpath)
	if err != nil {
		return err
	}
	defer closeDir(srcDir)

	destDir, err := openDir(dirpath)
	if err != nil {
		return err
	}
	defer closeDir(destDir)

	var args C.struct_btrfs_ioctl_vol_args_v2
	args.fd = C.__s64(getDirFd(srcDir))
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(destDir), C.BTRFS_IOC_SNAP_CREATE_V2,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to create btrfs snapshot: %v", errno.Error())
	}
	return nil
}

// IsSubvolume tests whether path `dirpath` is a btrfs subvolume
func IsSubvolume(dirpath string) (bool, error) {
	var bufStat syscall.Stat_t
	if err := syscall.Lstat(dirpath, &bufStat); err != nil {
		return false, err
	}

	// return true if it is a btrfs subvolume
	return bufStat.Ino == C.BTRFS_FIRST_FREE_OBJECTID, nil
}

func SubvolDelete(dirpath, name string) error {
	dir, err := openDir(dirpath)
	if err != nil {
		return err
	}
	defer closeDir(dir)

	var args C.struct_btrfs_ioctl_vol_args

	// walk the btrfs subvolumes
	walkSubvolumes := func(p string, f os.FileInfo, err error) error {
		// we want to check children only so skip itself
		// it will be removed after the filepath walk anyways
		if f.IsDir() && p != path.Join(dirpath, name) {
			sv, err := IsSubvolume(p)
			if err != nil {
				return fmt.Errorf("Failed to test if %s is a btrfs subvolume: %v", p, err)
			}
			if sv {
				if err := SubvolDelete(p, f.Name()); err != nil {
					return fmt.Errorf("Failed to destroy btrfs child subvolume (%s) of parent (%s): %v", p, dirpath, err)
				}
			}
		}
		return nil
	}
	if err := filepath.Walk(path.Join(dirpath, name), walkSubvolumes); err != nil {
		return fmt.Errorf("Recursively walking subvolumes for %s failed: %v", dirpath, err)
	}

	// all subvolumes have been removed
	// now remove the one originally passed in
	for i, c := range []byte(name) {
		args.name[i] = C.char(c)
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_SNAP_DESTROY,
		uintptr(unsafe.Pointer(&args)))
	if errno != 0 {
		return fmt.Errorf("Failed to destroy btrfs snapshot %s for %s: %v", dirpath, name, errno.Error())
	}
	return nil
}
