// +build !windows,linux,!darwin

package core

import (
	"syscall"

	"github.com/rs/zerolog/log"
)

func unmount(path string) error {
	err := syscall.Unmount(path, 0)

	log.Debug().Err(err).Msg("umount - linux")

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
