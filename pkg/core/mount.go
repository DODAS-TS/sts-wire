package core

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/DODAS-TS/sts-wire/pkg/rclone"
	"github.com/rs/zerolog/log"
)

const (
	numCheckAttempts = 10
	checkExeWait     = 1 * time.Second
	waitFileBusy     = 1 * time.Second
	closeChannelWait = 2 * time.Second
)

func CacheDir() (string, error) {
	cacheDir, errCacheDir := os.UserCacheDir()
	if errCacheDir != nil {
		log.Err(errCacheDir).Msg("Caching dir not available")

		return "", fmt.Errorf("CacheDir %w", errCacheDir)
	}

	return filepath.Join(cacheDir, "sts-wire"), nil
}

func ExePath() (string, error) {
	cacheDir, errCacheDir := CacheDir()
	if errCacheDir != nil {
		return "", errCacheDir
	}

	return filepath.Join(cacheDir, rcloneExeString), nil
}

func CheckExeFile(rcloneFile string, originalData []byte) error {
	var rcloneExeFile bytes.Buffer

	rcloneFileReader, errOpen := os.Open(rcloneFile)
	if errOpen != nil {
		log.Err(errOpen).Msg("Cannot open rclone executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errOpen)
	}

	defer rcloneFileReader.Close()

	// Check if it has executable permissions
	if runtime.GOOS != "windows" {
		rcloneInfo, err := rcloneFileReader.Stat()
		if err != nil {
			return fmt.Errorf("cannot have rclone file stat")
		}

		if rcloneInfo.Mode() != os.FileMode(exeFileMode) {
			return fmt.Errorf("rclone is not an executable")
		}
	}

	// Verify checksum
	_, errRead := rcloneExeFile.ReadFrom(rcloneFileReader)
	if errRead != nil {
		log.Err(errRead).Msg("Cannot read rclone executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errRead)
	}

	curExe := md5.New() //nolint:gosec

	_, errWriteCurExe := io.Copy(curExe, bytes.NewReader(rcloneExeFile.Bytes()))
	if errWriteCurExe != nil {
		log.Err(errWriteCurExe).Msg("Cannot calculate md5 for rclone executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errWriteCurExe)
	}

	rcloneExe := md5.New() //nolint:gosec

	_, errWriteRcloneExe := io.Copy(rcloneExe, bytes.NewReader(originalData))
	if errWriteRcloneExe != nil {
		log.Err(errWriteRcloneExe).Msg("Cannot calculate md5 for original rclone")

		return fmt.Errorf("prepare rclone %w", errWriteRcloneExe)
	}

	rcloneMd5 := curExe.Sum(nil)
	originalMd5 := rcloneExe.Sum(nil)
	log.Debug().Str("rcloneMd5",
		hex.EncodeToString(rcloneMd5)).Str("originalMd5",
		hex.EncodeToString(originalMd5)).Msg("rclone - executable checksum")

	if !bytes.Equal(rcloneMd5, originalMd5) {
		log.Err(nil).Msg("checksum not equal")

		return fmt.Errorf("checksum not equal")
	}

	return nil
}

func PrepareRclone() error { // nolint: funlen,gocognit
	baseDir, errCacheDir := CacheDir()
	if errCacheDir != nil {
		return errCacheDir
	}

	log.Debug().Str("basedir", baseDir).Msg("rclone")

	rcloneFile, errExePath := ExePath()
	if errExePath != nil {
		return errExePath
	}

	log.Debug().Str("rcloneFile", rcloneFile).Msg("rclone")
	log.Debug().Msg("rclone - get asset data")

	data := rclone.Executable
	log.Debug().Int("assetLen", len(data)).Msg("rclone")

	log.Debug().Msg("rclone - create executable base directories")

	errMkdir := os.MkdirAll(baseDir, os.ModePerm)
	if errMkdir != nil && !os.IsExist(errMkdir) {
		log.Err(errMkdir).Msg("Cannot create rclone cache dir")

		return fmt.Errorf("prepare rclone %w", errMkdir)
	}

	_, errStat := os.Stat(rcloneFile)
	log.Debug().Bool("missingExe", os.IsNotExist(errStat)).Msg("rclone")

	if errStat != nil && os.IsNotExist(errStat) { //nolint:nestif
		log.Debug().Msg("rclone - creating executable")

		rcloneExeFile, errCreate := os.OpenFile(rcloneFile, os.O_RDWR|os.O_CREATE|os.O_SYNC, fileMode)
		if errCreate != nil {
			log.Err(errCreate).Msg("rclone - creating executable - Cannot open with write permission the rclone executable in cache dir")

			// file busy
			if strings.Contains(errCreate.Error(), "file busy") {
				var errCheck error

				for attempt := 0; attempt < numCheckAttempts; attempt++ {
					log.Debug().Int("attempt", attempt).Msg("rclone - verify executable")

					errCheck = CheckExeFile(rcloneFile, data)
					if errCheck != nil {
						log.Err(errCheck).Int("attempt", attempt).Msg("rclone - verify executable - Cannot verify rclone executable in cache dir when file is busy")
					} else {
						break
					}

					time.Sleep(waitFileBusy)
				}

				if errCheck != nil {
					return errCheck
				}

				return nil
			}

			log.Err(errCreate).Msg("rclone - creating executable - Cannot create rclone executable in cache dir")

			return fmt.Errorf("prepare rclone %w", errCreate)
		}

		buff := bytes.NewReader(data)
		writtenData, errWrite := io.Copy(rcloneExeFile, buff)

		log.Debug().Int64("writtenData", writtenData).Msg("rclone")

		if errWrite != nil {
			log.Err(errWrite).Msg("rclone - creating executable - Cannot write rclone executable in cache dir")

			return fmt.Errorf("prepare rclone %w", errWrite)
		}

		rcloneExeFile.Close()

		log.Debug().Msg("rclone - change executable mod")

		errChmod := os.Chmod(rcloneFile, os.FileMode(exeFileMode))
		if errChmod != nil {
			log.Err(errChmod).Msg("rclone - creating executable - Cannot make rclone an executable in cache dir")

			return fmt.Errorf("prepare rclone %w", errChmod)
		}
	}

	var errCheck error

	for attempt := 0; attempt < numCheckAttempts; attempt++ {
		log.Debug().Int("attempt", attempt).Msg("rclone - verify executable")

		errCheck = CheckExeFile(rcloneFile, data)
		if errCheck != nil {
			log.Err(errCheck).Int("attempt", attempt).Msg("Cannot verify rclone executable in cache dir")
		} else {
			break
		}

		time.Sleep(checkExeWait)
	}

	if errCheck != nil {
		return fmt.Errorf("prepare rclone %w", errCheck)
	}

	return nil
}

func MountVolume(instance string, remotePath string, localPath string, configPath string) (*exec.Cmd, chan error, string, error) { // nolint: funlen,gocognit,lll
	log.Debug().Str("action", "prepare rclone").Msg("rclone - mount")

	errPrepare := PrepareRclone()
	if errPrepare != nil {
		log.Err(errPrepare).Msg("rclone - mount")

		return nil, nil, "", errPrepare
	}

	log.Debug().Str("action", "get file path").Msg("rclone - mount")

	rcloneFile, errExePath := ExePath()
	if errExePath != nil {
		return nil, nil, "", errExePath
	}

	log.Debug().Str("action", "make local dir").Msg("rclone - mount")

	if runtime.GOOS != "windows" {
		_, errLocalPath := os.Stat(localPath)
		if os.IsNotExist(errLocalPath) {
			errMkdir := os.MkdirAll(localPath, os.ModePerm)
			if errMkdir != nil {
				panic(errMkdir)
			}
		}
	}

	conf := fmt.Sprintf("%s:%s", instance, remotePath)

	log.Debug().Str("action", "prepare mounting points").Msg("rclone - mount")

	configPathAbs, errConfigPath := filepath.Abs(filepath.Join(configPath, "/rclone.conf"))
	if errConfigPath != nil {
		log.Err(errConfigPath).Msg("server")

		return nil, nil, "", fmt.Errorf("rclone config abs: %w", errConfigPath)
	}

	logPath, errLogPath := filepath.Abs(filepath.Join(configPath, "/rclone.log"))
	if errLogPath != nil {
		log.Err(errLogPath).Msg("server")

		return nil, nil, "", fmt.Errorf("rclone log abs: %w", errLogPath)
	}

	localPathAbs, errLocalPath := filepath.Abs(localPath)
	if errLocalPath != nil {
		log.Err(errLocalPath).Msg("server")

		return nil, nil, "", fmt.Errorf("local path abs: %w", errLocalPath)
	}

	log.Debug().Str("command", strings.Join([]string{
		rcloneFile,
		"--config",
		configPathAbs,
		"--no-check-certificate",
		"mount",
		// "--daemon",
		"--log-file",
		logPath,
		"--log-level=DEBUG",
		"--vfs-cache-mode",
		"full",
		"--no-modtime",
		conf,
		localPathAbs,
	}, " ")).Msg("rclone - mount")

	rcloneCmd := exec.Command(
		rcloneFile,
		"--config",
		configPathAbs,
		"--no-check-certificate",
		"mount",
		// "--daemon",
		"--log-file",
		logPath,
		"--log-level=DEBUG",
		"--vfs-cache-mode",
		"full",
		"--no-modtime",
		conf,
		localPathAbs,
	)

	log.Debug().Str("action", "start rclone").Msg("rclone - mount")

	errStart := rcloneCmd.Start()
	if errStart != nil {
		panic(errStart)
	}

	cmdErr := make(chan error)

	cmdErrorCheck := func() {
		procState, errWait := rcloneCmd.Process.Wait()

		// Wait for main server to close channel if User pressed Ctrl+C
		time.Sleep(closeChannelWait)

		openChannel := true
		// Check if channel still is open
		select {
		case _, ok := <-cmdErr:
			if !ok {
				openChannel = false
			}
		default:
		}

		if procState != nil {
			// ref: https://rclone.org/docs/#exit-code
			switch exitCode := procState.ExitCode(); exitCode {
			case 0:
				log.Debug().Int("exitCode",
					exitCode).Msg("success")
			case 1:
				log.Debug().Int("exitCode",
					exitCode).Msg("Syntax or usage error")
			case 2:
				log.Debug().Int("exitCode",
					exitCode).Msg("Error not otherwise categorised")
			case 3:
				log.Debug().Int("exitCode",
					exitCode).Msg("Directory not found")
			case 4:
				log.Debug().Int("exitCode",
					exitCode).Msg("File not found")
			case 5:
				log.Debug().Int("exitCode",
					exitCode).Msg("Temporary error (one that more retries might fix) (Retry errors)")
			case 6:
				log.Debug().Int("exitCode",
					exitCode).Msg("Less serious errors (like 461 errors from dropbox) (NoRetry errors)")
			case 7:
				log.Debug().Int("exitCode",
					exitCode).Msg("Fatal error (one that more retries won't fix, like account suspended) (Fatal errors)")
			case 8:
				log.Debug().Int("exitCode",
					exitCode).Msg("Transfer exceeded - limit set by --max-transfer reached")
			case 9:
				log.Debug().Int("exitCode",
					exitCode).Msg("Operation successful, but no files transferred")
			}
		}

		if openChannel { //nolint:nestif
			// rclone exited with errors
			if errWait != nil {
				cmdErr <- errWait
			}

			defer close(cmdErr)
		} else {
			if errWait == nil {
				// rclone not exited after user pressed Ctrl+C
				if !procState.Exited() {
					panic("rclone termination error")
				}
			} else if errWait.Error() != "wait: no child processes" {
				// rclone exited for unknown reason
				panic(errWait)
			}
		}
	}
	go cmdErrorCheck()

	return rcloneCmd, cmdErr, logPath, nil
}

func RcloneLogErrors(logPath string) chan string {
	outErrors := make(chan string)

	go func() {
		defer close(outErrors)

		latestErrors := make([]string, 0)

		readFile, err := os.Open(logPath)
		if err != nil {
			log.Err(err).Str("logFile", logFile).Msg("failed to open log file")
		}

		defer readFile.Close()

		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			curLine := fileScanner.Text()

			switch {
			case strings.Contains(curLine, "INFO") && strings.Contains(curLine, "Exiting..."):
				latestErrors = make([]string, 0)
			case strings.Contains(curLine, "error"), strings.Contains(curLine, "ERROR"):
				latestErrors = append(latestErrors, curLine)
			}
		}

		for _, foundErr := range latestErrors {
			outErrors <- foundErr
		}
	}()

	return outErrors
}
