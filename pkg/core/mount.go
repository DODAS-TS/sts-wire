package core

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/DODAS-TS/sts-wire/pkg/rclone"
	"github.com/rs/zerolog/log"
)

func CacheDir() (string, error) {
	cacheDir, errCacheDir := os.UserCacheDir()
	if errCacheDir != nil {
		log.Err(errCacheDir).Msg("Caching dir not available")

		return "", fmt.Errorf("CacheDir %w", errCacheDir)
	}

	return path.Join(cacheDir, "sts-wire"), nil
}

func ExePath() (string, error) {
	cacheDir, errCacheDir := CacheDir()
	if errCacheDir != nil {
		return "", errCacheDir
	}

	return path.Join(cacheDir, "rclone"), nil
}

func PrepareRclone() error { // nolint: funlen
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

	data, errAsset := rclone.Asset("rclone")
	log.Debug().Int("assetLen", len(data)).Msg("rclone")

	if errAsset != nil {
		log.Err(errAsset).Msg("Rclone asset for Darwin not found")

		return fmt.Errorf("prepare rclone %w", errAsset)
	}

	log.Debug().Msg("rclone - create executable")

	errMkdir := os.MkdirAll(baseDir, os.ModePerm)
	if errMkdir != nil && !os.IsExist(errMkdir) {
		log.Err(errMkdir).Msg("Cannot create rclone cache dir")

		return fmt.Errorf("prepare rclone %w", errMkdir)
	}

	rcloneExeFile, errCreate := os.OpenFile(rcloneFile, os.O_RDWR|os.O_CREATE|os.O_SYNC, fileMode)
	if errCreate != nil {
		log.Err(errCreate).Msg("Cannot create rclone executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errCreate)
	}

	buff := bytes.NewReader(data)
	writtenData, errWrite := io.Copy(rcloneExeFile, buff)

	log.Debug().Int64("writtenData", writtenData).Msg("rclone")

	if errWrite != nil {
		log.Err(errWrite).Msg("Cannot write rclone executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errWrite)
	}

	rcloneExeFile.Close()

	log.Debug().Msg("rclone - change executable mod")

	errChmod := os.Chmod(rcloneFile, os.FileMode(exeFileMode))
	if errChmod != nil {
		log.Err(errChmod).Msg("Cannot make rclone an executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errChmod)
	}

	return nil
}

func MountVolume(instance string, remotePath string, localPath string, configPath string) (*exec.Cmd, chan error, string, error) { // nolint: funlen
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

	_, errLocalPath := os.Stat(localPath)
	if os.IsNotExist(errLocalPath) {
		errMkdir := os.MkdirAll(localPath, os.ModePerm)
		if errMkdir != nil {
			panic(errMkdir)
		}
	}

	conf := fmt.Sprintf("%s:%s", instance, remotePath)

	log.Debug().Str("action", "prepare mounting points").Msg("rclone - mount")

	configPathAbs, errConfigPath := filepath.Abs(path.Join(configPath, "/rclone.conf"))
	if errConfigPath != nil {
		log.Err(errConfigPath).Msg("server")

		return nil, nil, "", errConfigPath
	}

	logPath, errLogPath := filepath.Abs(path.Join(configPath, "/rclone.log"))
	if errLogPath != nil {
		log.Err(errLogPath).Msg("server")

		return nil, nil, "", errLogPath
	}

	localPathAbs, errLocalPath := filepath.Abs(localPath)
	if errLocalPath != nil {
		log.Err(errLocalPath).Msg("server")

		return nil, nil, "", errLocalPath
	}

	log.Debug().Str("command", strings.Join([]string{
		rcloneFile,
		"--config",
		configPathAbs,
		"--no-check-certificate",
		"mount",
		//"--daemon",
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
		//"--daemon",
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
		time.Sleep(2 * time.Second)

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

		readFile, err := os.Open(logPath)

		if err != nil {
			log.Err(err).Str("logFile", logFile).Msg("failed to open log file")
		}

		defer readFile.Close()

		fileScanner := bufio.NewScanner(readFile)
		fileScanner.Split(bufio.ScanLines)

		for fileScanner.Scan() {
			curLine := fileScanner.Text()
			if strings.Contains(curLine, "error") {
				outErrors <- fileScanner.Text()
			}
		}
	}()

	return outErrors
}
