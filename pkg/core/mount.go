package core

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

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

	return path.Join(cacheDir, "rclone_osx"), nil
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

	rcloneExeFile, errCreate := os.OpenFile(rcloneFile, os.O_RDWR|os.O_CREATE, fileMode)
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

func MountVolume(instance string, remotePath string, localPath string, configPath string) (*exec.Cmd, error) { // nolint: funlen
	log.Debug().Str("action", "prepare rclone").Msg("rclone - mount")

	errPrepare := PrepareRclone()
	if errPrepare != nil {
		log.Err(errPrepare).Msg("rclone - mount")

		return nil, errPrepare
	}

	log.Debug().Str("action", "get file path").Msg("rclone - mount")

	rcloneFile, errExePath := ExePath()
	if errExePath != nil {
		return nil, errExePath
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

		return nil, errConfigPath
	}

	logPath, errLogPath := filepath.Abs(path.Join(configPath, "/rclone.log"))
	if errLogPath != nil {
		log.Err(errLogPath).Msg("server")

		return nil, errLogPath
	}

	localPathAbs, errLocalPath := filepath.Abs(localPath)
	if errLocalPath != nil {
		log.Err(errLocalPath).Msg("server")

		return nil, errLocalPath
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

	return rcloneCmd, nil
}
