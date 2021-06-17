// +build !windows,!linux, darwin

package core

import "syscall"

func unmount(path string) error {
	err := syscall.Unmount(path, 0)
	if err != nil {
		if err != syscall.EINVAL {
			return err
		} else {
			return nil // syscall.EINVAL for invalid flag (because it is not a mount point)
		}
	}

	return nil
}
