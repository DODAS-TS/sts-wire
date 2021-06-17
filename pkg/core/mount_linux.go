// +build !windows,linux,!darwin

package core

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

func unmount(path string) error {
	err := syscall.Unmount(path, 0)
	if err != nil {
		if err != syscall.EINVAL {
			return err
		} else {
			// syscall.EINVAL target is not a mount point.
			// - https://man7.org/linux/man-pages/man2/umount.2.html
			return nil
		}
	}

	return nil
}
