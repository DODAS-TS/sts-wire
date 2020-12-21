package validator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	ErrNoValidPath = errors.New("no valid path")
	validPath      = regexp.MustCompile(`^([a-z_\-\s0-9\.\/]+)+$`)
)

func LocalPath(localMountPoint string) (bool, error) {
	if !validPath.MatchString(localMountPoint) {
		return false, ErrNoValidPath
	}

	absPath, errAbs := filepath.Abs(filepath.Clean(localMountPoint))
	if errAbs != nil {
		return false, fmt.Errorf("no valid path: %w", errAbs)
	}

	_, errStat := os.Stat(absPath)
	if errStat != nil && !os.IsNotExist(errStat) {
		return false, fmt.Errorf("no valid path: %w", errStat)
	}

	errMkdir := os.MkdirAll(absPath, os.ModePerm)
	if errMkdir != nil {
		return false, fmt.Errorf("no valid path: %w", errStat)
	}

	os.RemoveAll(absPath)

	parentFolder, errStat := os.Stat(filepath.Dir(absPath))
	if errStat != nil {
		return false, fmt.Errorf("no valid path: %w", errStat)
	}

	if !parentFolder.IsDir() || os.IsNotExist(errStat) {
		return false, fmt.Errorf("%w: %s is not valid because it has not a parent folder",
			ErrNoValidPath, localMountPoint,
		)
	}

	return true, nil
}
