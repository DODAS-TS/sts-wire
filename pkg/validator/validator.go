package validator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var (
	ErrNoValidFile         = errors.New("no valid file")
	validFile              = regexp.MustCompile(`^([a-zA-Z_\-\s0-9\.\/]+)+$`)
	ErrNoValidLogFile      = errors.New("no valid log file")
	validLogFile           = regexp.MustCompile(`^([a-zA-Z_\-\s0-9\.\/]+)+(.log)$`)
	ErrNoValidPath         = errors.New("no valid path")
	validPath              = regexp.MustCompile(`^([a-z_\-\s0-9\.\/]+)+$`)
	ErrNoValidEndpoint     = errors.New("no valid s3 endpoint")
	ErrNoValidWebURL       = errors.New("no valid web URL")
	validWebURL            = regexp.MustCompile(`^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)?(?:localhost)?[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`)
	ErrNoValidS3RemotePath = errors.New("no valid s3 remote path")
	validRemotePath        = regexp.MustCompile(`^/[\w\-_/]+$`)
	ErrNoValidInstanceName = errors.New("no valid instance name")
	validInstanceName      = regexp.MustCompile(`^[\w\-_]+$`)
)

func LogFile(logFilePath string) (bool, error) {
	if !validFile.MatchString(logFilePath) {
		return false, ErrNoValidFile
	}

	if !validLogFile.MatchString(logFilePath) {
		return false, ErrNoValidLogFile
	}

	_, errAbs := filepath.Abs(filepath.Clean(logFilePath))
	if errAbs != nil {
		return false, fmt.Errorf("no valid path: %w", errAbs)
	}

	return true, nil
}

func InstanceName(url string) (bool, error) {
	if !validInstanceName.MatchString(url) {
		return false, ErrNoValidInstanceName
	}

	return true, nil
}

func WebURL(url string) (bool, error) {
	if !validWebURL.MatchString(url) {
		return false, ErrNoValidWebURL
	}

	return true, nil
}

func S3Endpoint(endpoint string) (bool, error) {
	if valid, err := WebURL(endpoint); err != nil {
		return valid, ErrNoValidEndpoint
	}

	return true, nil
}

func RemotePath(remotePath string) (bool, error) {
	if !validRemotePath.MatchString(remotePath) {
		return false, ErrNoValidS3RemotePath
	}

	return true, nil
}

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
