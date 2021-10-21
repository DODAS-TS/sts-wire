// +build !windows,linux darwin

package core

import (
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"golang.org/x/sys/unix"
)

const (
	rcloneExeString = "rclone"
)

// checkMountpoint verify if a path is a mount point.
// Note: this is a laxy check because it not detects bind mounts.
func checkMountpoint(path string) (bool, error) {
	var st unix.Stat_t

	if err := unix.Lstat(path, &st); err != nil {
		if err == unix.ENOENT {
			// ENOENT -> not a mount point
			return false, nil
		}

		return false, &os.PathError{Op: "stat", Path: path, Err: err}
	}

	dev := st.Dev

	parent := filepath.Dir(path)
	if err := unix.Lstat(parent, &st); err != nil {
		return false, &os.PathError{Op: "stat", Path: parent, Err: err}
	}

	log.Debug().Int("dev", int(dev)).Int("parent dev", int(st.Dev)).Msg("checkMountpoint")

	if dev != st.Dev {
		// If the Device differs from that of parent, it is a mount point
		return true, nil
	}

	return false, nil
}
