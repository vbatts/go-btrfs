// +build linux,!btrfs_noversion

package btrfs

/*
#include <btrfs/version.h>

// around version 3.16, they did not define lib version yet
#ifndef BTRFS_LIB_VERSION
#define BTRFS_LIB_VERSION -1
#endif

// upstream had removed it, but now it will be coming back
#ifndef BTRFS_BUILD_VERSION
#define BTRFS_BUILD_VERSION "-"
#endif
*/
import "C"

// BuildVersion returns the build version of libbtrfs, if available
func BuildVersion() string {
	return string(C.BTRFS_BUILD_VERSION)
}

// LibVersion returns the library version of libbtrfs, if available
func LibVersion() int {
	return int(C.BTRFS_LIB_VERSION)
}
