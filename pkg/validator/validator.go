package validator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	minRefreshTokenDuration = 15
)

var (
	ErrNoValidFile                  = errors.New("no valid file")
	validFile                       = regexp.MustCompile(`^([a-zA-Z_\:\\\-\s0-9\.\/]+)+$`)
	ErrNoValidLogFile               = errors.New("no valid log file")
	validLogFile                    = regexp.MustCompile(`^([a-zA-Z_\:\\\-\s0-9\.\/]+)+(\.log)$`)
	ErrNoValidPath                  = errors.New("no valid path")
	validPath                       = regexp.MustCompile(`^([a-z_\-\s0-9\.\/]+)+$`)
	ErrNoValidEndpoint              = errors.New("no valid s3 endpoint")
	ErrNoValidWebURL                = errors.New("no valid web URL")
	validWebURL                     = regexp.MustCompile(`^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)?(?:localhost)?[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`)
	ErrNoValidS3RemotePath          = errors.New("no valid s3 remote path")
	validRemotePath                 = regexp.MustCompile(`^/[\w\-_/]+$`)
	ErrNoValidInstanceName          = errors.New("no valid instance name")
	validInstanceName               = regexp.MustCompile(`^[\w\-_]+$`)
	ErrNoValidRefreshTokenRenewTime = errors.New("no valid refresh token time duration: min 15min")
)

// RefreshTokenRenew checks if the number of minutes are valid: minimum is 15min.
func RefreshTokenRenew(minutes int) (bool, error) {
	if minutes < minRefreshTokenDuration {
		return false, ErrNoValidRefreshTokenRenewTime
	}

	return true, nil
}

// LogFile checks if the path indicated for the log file is valid.
func LogFile(logFilePath string) (bool, error) {
	if !validLogFile.MatchString(logFilePath) {
		return false, ErrNoValidLogFile
	}

	return true, nil
}

// InstanceName checks if the user instance name is valid.
func InstanceName(url string) (bool, error) {
	if !validInstanceName.MatchString(url) {
		return false, ErrNoValidInstanceName
	}

	return true, nil
}

// WebURL checks if the given URL is a valid Web URL: http://example.com:port
func WebURL(url string) (bool, error) {
	if !validWebURL.MatchString(url) {
		return false, ErrNoValidWebURL
	}

	return true, nil
}

// S3Endpoint checks if the given endpoint is a valid s3 endpoint.
func S3Endpoint(endpoint string) (bool, error) {
	if valid, err := WebURL(endpoint); err != nil {
		return valid, ErrNoValidEndpoint
	}

	// The ending slash is not required
	// reference: https://github.com/minio/minio/blob/master/docs/sts/web-identity.md#sample-post-request
	// if endpoint[len(endpoint)-1] != '/' {
	// 	return false, ErrNoValidEndpoint
	// }

	return true, nil
}

// RemotePath checks if the remote path is valid (a valid bucket path: /something...).
func RemotePath(remotePath string) (bool, error) {
	if !validRemotePath.MatchString(remotePath) {
		return false, ErrNoValidS3RemotePath
	}

	return true, nil
}

// LocalPath verify if the path indicated could be a proper mountpoint.
func LocalPath(localMountPoint string) (bool, error) {
	if localMountPoint == "/" {
		return false, ErrNoValidPath
	}

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

	// TODO: better check over folder permissions
	// errMkdir := os.MkdirAll(absPath, os.ModePerm)
	// if errMkdir != nil {
	// 	return false, fmt.Errorf("no valid path: %w", errStat)
	// }

	// os.RemoveAll(absPath)

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
