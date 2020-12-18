package validator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrNoValidPath = errors.New("no valid path")

func LocalPath(localMountPoint string) (bool, error) {
	absPath, errAbs := filepath.Abs(filepath.Clean(localMountPoint))
	if errAbs != nil {
		return false, fmt.Errorf("no valid path: %w", errAbs)
	}

	targetFolder, errStat := os.Stat(absPath)
	if errStat != nil && !os.IsNotExist(errStat) {
		return false, fmt.Errorf("no valid path: %w", errStat)
	}

	if os.IsNotExist(errStat) {
		errMkdir := os.Mkdir(absPath, os.ModePerm)
		if errMkdir != nil {
			return false, fmt.Errorf("no valid path: %w", errStat)
		}

		os.RemoveAll(absPath)
	} else if !targetFolder.IsDir() {
		return false, ErrNoValidPath
	}

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
