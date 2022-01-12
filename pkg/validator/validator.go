package validator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	minRefreshTokenDuration = 15
)

var (
	ErrNoValidSize                  = errors.New("no valid uid")
	validSize                       = regexp.MustCompile(`^((\d*\.?\d+)[KMBkmb]?){1}$`)
	ErrNoValidUID                   = errors.New("no valid uid")
	validUID                        = regexp.MustCompile(`^([\d]+){4}$`)
	ErrNoValidPermission            = errors.New("no valid permission")
	validPermission                 = regexp.MustCompile(`^([\d]+){4}$`)
	ErrNoValidDuration              = errors.New("no valid duration")
	validDuration                   = regexp.MustCompile(`^(\d+h){0,1}(\d+(m|ms){0,1}){0,1}(\d+s){0,1}$`)
	ErrNoValidFile                  = errors.New("no valid file")
	validFile                       = regexp.MustCompile(`^([a-zA-Z_\:\\\-\s0-9\.\/]+)+$`)
	ErrNoValidLogFile               = errors.New("no valid log file")
	validLogFile                    = regexp.MustCompile(`^([a-zA-Z_\:\\\-\s0-9\.\/]+)+(\.log)$`)
	ErrNoValidPath                  = errors.New("no valid path")
	validPath                       = regexp.MustCompile(`^([a-zA-Z_\-\s0-9\.\/]+)+$`)
	ErrNoValidEndpoint              = errors.New("no valid s3 endpoint")
	ErrNoValidWebURL                = errors.New("no valid web URL")
	validWebURL                     = regexp.MustCompile(`^(?:http(s)?:\/\/)?[\w.-]+(?:\.[\w\.-]+)?(?:localhost)?[\w\-\._~:/?#[\]@!\$&'\(\)\*\+,;=.]+$`)
	ErrNoValidS3RemotePath          = errors.New("no valid s3 remote path")
	validRemotePath                 = regexp.MustCompile(`^/[\w\-_/]+$`)
	ErrNoValidInstanceName          = errors.New("no valid instance name")
	validInstanceName               = regexp.MustCompile(`^[\w\-_]+$`)
	ErrNoValidRefreshTokenRenewTime = errors.New("no valid refresh token time duration: min 15min")
	ErrNoValidRcloneMountOption     = errors.New("mount option not valid")
	rcloneMountOptions              map[string]interface{} // nolint:gochecknoglobals
)

func init() {
	rcloneMountOptions = make(map[string]interface{})
	// Allow mounting over a non-empty directory. Not supported on Windows.
	rcloneMountOptions["--allow-non-empty"] = nil
	// Allow access to other users. Not supported on Windows.
	rcloneMountOptions["--allow-other"] = nil
	// Allow access to root user. Not supported on Windows.
	rcloneMountOptions["--allow-root"] = nil
	// Use asynchronous reads. Not supported on Windows. (default true)
	rcloneMountOptions["--async-read"] = nil
	// Time for which file/directory attributes are cached. (default 1s)
	rcloneMountOptions["--attr-timeout"] = validDuration
	// Run mount as a daemon (background mode). Not supported on Windows.
	rcloneMountOptions["--daemon"] = nil
	// Time limit for rclone to respond to kernel. Not supported on Windows.
	rcloneMountOptions["--daemon-timeout"] = validDuration
	// Debug the FUSE internals - needs -v.
	rcloneMountOptions["--debug-fuse"] = nil
	// Makes kernel enforce access control based on the file mode. Not supported on Windows.
	rcloneMountOptions["--default-permissions"] = nil
	// Time to cache directory entries for. (default 5m0s)
	rcloneMountOptions["--dir-cache-time"] = validDuration
	// Directory permissions (default 0777)
	rcloneMountOptions["--dir-perms"] = validPermission
	// File permissions (default 0666)
	rcloneMountOptions["--file-perms"] = validPermission
	// Flags or arguments to be passed direct to libfuse/WinFsp. Repeat if required.
	// rcloneMountOptions["--fuse-flag stringArray"] = []string{}
	// Override the gid field set by the filesystem. Not supported on Windows. (default 1000)
	rcloneMountOptions["--gid"] = validUID
	// The number of bytes that can be prefetched for sequential reads. Not supported on Windows. (default 128k)
	rcloneMountOptions["--max-read-ahead"] = validSize
	// Mount as remote network drive, instead of fixed disk drive. Supported on Windows only
	rcloneMountOptions["--network-mode"] = nil
	// Don't compare checksums on up/download.
	rcloneMountOptions["--no-checksum"] = nil
	// Don't read/write the modification time (can speed things up).
	rcloneMountOptions["--no-modtime"] = nil
	// Don't allow seeking in files.
	rcloneMountOptions["--no-seek"] = nil
	// Ignore Apple Double (._) and .DS_Store files. Supported on OSX only. (default true)
	rcloneMountOptions["--noappledouble"] = nil
	// Ignore all "com.apple.*" extended attributes. Supported on OSX only.
	rcloneMountOptions["--noapplexattr"] = nil
	// Option for libfuse/WinFsp. Repeat if required.
	rcloneMountOptions["--option stringArray"] = nil
	// Time to wait between polling for changes. Must be smaller than dir-cache-time. Only on supported remotes. Set to 0 to disable. (default 1m0s)
	rcloneMountOptions["--poll-interval"] = validDuration
	// Mount read-only.
	rcloneMountOptions["--read-only"] = nil
	// Override the uid field set by the filesystem. Not supported on Windows. (default 1000)
	rcloneMountOptions["--uid"] = validUID
	// Override the permission bits set by the filesystem. Not supported on Windows.
	rcloneMountOptions["--umask int"] = nil
	// Max age of objects in the cache. (default 1h0m0s)
	rcloneMountOptions["--vfs-cache-max-age"] = validDuration
	// Max total size of objects in the cache. (default off)
	rcloneMountOptions["--vfs-cache-max-size"] = validSize
	// Cache mode off|minimal|writes|full (default off)
	rcloneMountOptions["--vfs-cache-mode"] = []string{"off", "minimal", "writes", "full"}
	// Interval to poll the cache for stale objects. (default 1m0s)
	rcloneMountOptions["--vfs-cache-poll-interval"] = validDuration
	// If a file name not found, find a case insensitive match.
	rcloneMountOptions["--vfs-case-insensitive"] = nil
	// Extra read ahead over --buffer-size when using cache-mode full.
	rcloneMountOptions["--vfs-read-ahead"] = validSize
	// Read the source objects in chunks. (default 128M)
	rcloneMountOptions["--vfs-read-chunk-size"] = validSize
	// If greater than --vfs-read-chunk-size, double the chunk size after each chunk read, until the limit is reached. 'off' is unlimited. (default off)
	rcloneMountOptions["--vfs-read-chunk-size-limit"] = validSize
	// Time to wait for in-sequence read before seeking. (default 20ms)
	rcloneMountOptions["--vfs-read-wait"] = validDuration
	// Use the rclone size algorithm for Used size.
	//rcloneMountOptions["--vfs-used-is-size rclone size"] = []string{}
	// Time to writeback files after last use when using cache. (default 5s)
	rcloneMountOptions["--vfs-write-back"] = validDuration
	// Time to wait for in-sequence write before giving error. (default 1s)
	rcloneMountOptions["--vfs-write-wait"] = validDuration
	// Set the volume name. Supported on Windows and OSX only.
	rcloneMountOptions["--volname"] = validPath
	// Makes kernel buffer writes before sending them to rclone.
	// Without this, writethrough caching is used. Not supported on Windows.
	rcloneMountOptions["--write-back-cache"] = nil
}

// RcloneLogLevel checks if the level is a valid rclone log level
func RcloneLogLevel(level string) (bool, error) {
	switch strings.ToLower(level) {
	case "error", "notice", "info", "debug":
		return true, nil
	}

	return false, nil
}

// RcloneMountFlags checks if the flags are valid rclone mount flags
func RcloneMountFlags(flagList string) (bool, error) {
	parts := strings.Split(flagList, " ")

	for idx := 0; idx < len(parts); idx++ {
		curPart := parts[idx]
		if value, inMap := rcloneMountOptions[curPart]; inMap {
			switch checkMethod := value.(type) {
			case nil:
				continue
			case *regexp.Regexp:
				nextPart := parts[idx+1]
				if !checkMethod.MatchString(nextPart) {
					return false, fmt.Errorf("%w '%s %s'", ErrNoValidRcloneMountOption, curPart, nextPart)
				}

				idx++
			case []string:
				nextPart := parts[idx+1]
				found := false

				for _, val := range value.([]string) {
					if nextPart == val {
						found = true
					}
				}

				if !found {
					return false, fmt.Errorf("%w '%s %s'", ErrNoValidRcloneMountOption, curPart, nextPart)
				}
				idx++
			}
		} else {
			return false, fmt.Errorf("%w '%s'", ErrNoValidRcloneMountOption, curPart)
		}
	}

	return true, nil
}

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
		return false, fmt.Errorf("no valid path: %w, abs error", errAbs)
	}

	_, errStat := os.Stat(absPath)
	if errStat != nil && !os.IsNotExist(errStat) {
		return false, fmt.Errorf("no valid path: %w, something wront", errStat)
	}

	// TODO: better check over folder permissions
	// errMkdir := os.MkdirAll(absPath, os.ModePerm)
	// if errMkdir != nil {
	// 	return false, fmt.Errorf("no valid path: %w", errStat)
	// }

	// os.RemoveAll(absPath)

	parentFolder, errStat := os.Stat(filepath.Dir(absPath))
	if errStat != nil {
		return false, fmt.Errorf("no valid path: %w, parent folder", errStat)
	}

	if !parentFolder.IsDir() || os.IsNotExist(errStat) {
		return false, fmt.Errorf("%w: %s is not valid because it has not a parent folder",
			ErrNoValidPath, localMountPoint,
		)
	}

	return true, nil
}
