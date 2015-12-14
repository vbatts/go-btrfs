// +build !linux

package btrfs

import "github.com/Sirupsen/logrus"

func init() {
	logrus.Errorf("btrfs is not supported on this platform")
}
