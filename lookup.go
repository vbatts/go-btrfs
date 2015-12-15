// +build linux

package btrfs

/*
#include <stdlib.h>
#include <btrfs/ioctl.h>
#include <btrfs/ctree.h>
*/
import "C"

import (
	"fmt"
	"syscall"
	"unsafe"
)

func newInodeLookupArgs(args C.struct_btrfs_ioctl_ino_lookup_args) *inodeLookupArgs {
	return &inodeLookupArgs{
		Name:     C.GoString((*C.char)(unsafe.Pointer(&args.name[0]))),
		TreeID:   uint64(args.treeid),
		ObjectID: uint64(args.objectid),
	}
}

type inodeLookupArgs struct {
	TreeID, ObjectID uint64
	Name             string
}

func (ila *inodeLookupArgs) C() C.struct_btrfs_ioctl_ino_lookup_args {
	var args C.struct_btrfs_ioctl_ino_lookup_args
	args.objectid = C.__u64(ila.ObjectID)
	args.treeid = C.__u64(ila.TreeID)
	if ila.Name != "" {
		str := [C.BTRFS_INO_LOOKUP_PATH_MAX]C.char{}
		for i := 0; i < len(ila.Name) && i < C.BTRFS_INO_LOOKUP_PATH_MAX; i++ {
			str[i] = C.char(ila.Name[i])
		}
		args.name = str
	}
	return args
}

func inodeLookup(dirpath string) (*inodeLookupArgs, error) {
	dir, err := openDir(dirpath)
	if err != nil {
		return nil, err
	}
	defer closeDir(dir)
	var inoArgs C.struct_btrfs_ioctl_ino_lookup_args
	inoArgs.objectid = C.BTRFS_FIRST_FREE_OBJECTID
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, getDirFd(dir), C.BTRFS_IOC_INO_LOOKUP,
		uintptr(unsafe.Pointer(&inoArgs)))
	if errno != 0 {
		return nil, fmt.Errorf("Failed to lookup btrfs inode: %v", errno.Error())
	}
	return newInodeLookupArgs(inoArgs), nil
}

func huurr(dirpath string) ([]string, error) {
	inoArgs, err := inodeLookup(dirpath)
	if err != nil {
		return nil, err
	}

	var searchKey C.struct_btrfs_ioctl_search_key
	searchKey.min_objectid = C.__u64(inoArgs.TreeID)
	searchKey.max_objectid = C.__u64(inoArgs.TreeID)
	searchKey.min_type = C.BTRFS_ROOT_ITEM_KEY
	searchKey.max_type = C.BTRFS_ROOT_ITEM_KEY
	searchKey.max_offset = (1<<48 - 1)
	searchKey.max_transid = (1<<48 - 1)

	return nil, nil
}
