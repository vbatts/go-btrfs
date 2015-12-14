// +build linux

/*
Initial source imported from github.com/docker/docker ./pkg/devicemapper/attach_loop.go @ 75d69ce0da2e360773736502acd92d4a9cf7faa5
See LICENSE.docker
*/

package loop

import (
	"fmt"
	"os"
	"syscall"

	"github.com/Sirupsen/logrus"
)

func stringToLoopName(src string) [LoNameSize]uint8 {
	var dst [LoNameSize]uint8
	copy(dst[:], src[:])
	return dst
}

func getNextFreeLoopbackIndex() (int, error) {
	f, err := os.OpenFile("/dev/loop-control", os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	index, err := ioctlLoopCtlGetFree(f.Fd())
	if index < 0 {
		index = 0
	}
	return index, err
}

func openNextAvailableLoopback(index int, sparseFile *os.File) (loopFile *os.File, err error) {
	// Start looking for a free /dev/loop
	for {
		target := fmt.Sprintf("/dev/loop%d", index)
		index++

		fi, err := os.Stat(target)
		if err != nil {
			if os.IsNotExist(err) {
				logrus.Errorf("There are no more loopback devices available.")
			}
			return nil, ErrAttachLoopbackDevice
		}

		if fi.Mode()&os.ModeDevice != os.ModeDevice {
			logrus.Errorf("Loopback device %s is not a block device.", target)
			continue
		}

		// OpenFile adds O_CLOEXEC
		loopFile, err = os.OpenFile(target, os.O_RDWR, 0644)
		if err != nil {
			logrus.Errorf("Error opening loopback device: %s", err)
			return nil, ErrAttachLoopbackDevice
		}

		// Try to attach to the loop file
		if err := ioctlLoopSetFd(loopFile.Fd(), sparseFile.Fd()); err != nil {
			loopFile.Close()

			// If the error is EBUSY, then try the next loopback
			if err != syscall.EBUSY {
				logrus.Errorf("Cannot set up loopback device %s: %s", target, err)
				return nil, ErrAttachLoopbackDevice
			}

			// Otherwise, we keep going with the loop
			continue
		}
		// In case of success, we finished. Break the loop.
		break
	}

	// This can't happen, but let's be sure
	if loopFile == nil {
		logrus.Errorf("Unreachable code reached! Error attaching %s to a loopback device.", sparseFile.Name())
		return nil, ErrAttachLoopbackDevice
	}

	return loopFile, nil
}

// AttachLoopDevice attaches the given sparse file to the next
// available loopback device. It returns an opened *os.File.
/*
 Example
```

fh, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0600)
if err != nil {
  return err
}

if err := fh.Truncate(size); err != nil {
  return err
}
if err := fh.Close(); err != nil {
  return err
}

lHandle, err := loop.AttachLoopDevice(filename)
if err != nil {
  return err
}

[...]

```
*/
func AttachLoopDevice(filename string) (loop *os.File, err error) {

	// Try to retrieve the next available loopback device via syscall.
	// If it fails, we discard error and start looping for a
	// loopback from index 0.
	startIndex, err := getNextFreeLoopbackIndex()
	if err != nil {
		logrus.Debugf("Error retrieving the next available loopback: %s", err)
	}

	// OpenFile adds O_CLOEXEC
	sparseFile, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		logrus.Errorf("Error opening sparse file %s: %s", filename, err)
		return nil, ErrAttachLoopbackDevice
	}
	defer sparseFile.Close()

	loopFile, err := openNextAvailableLoopback(startIndex, sparseFile)
	if err != nil {
		return nil, err
	}

	// Set the status of the loopback device
	loopInfo := &loopInfo64{
		loFileName: stringToLoopName(loopFile.Name()),
		loOffset:   0,
		loFlags:    LoFlagsAutoClear,
	}

	if err := ioctlLoopSetStatus64(loopFile.Fd(), loopInfo); err != nil {
		logrus.Errorf("Cannot set up loopback device info: %s", err)

		// If the call failed, then free the loopback device
		if err := ioctlLoopClrFd(loopFile.Fd()); err != nil {
			logrus.Errorf("Error while cleaning up the loopback device")
		}
		loopFile.Close()
		return nil, ErrAttachLoopbackDevice
	}

	return loopFile, nil
}
