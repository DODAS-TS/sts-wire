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

const (
	exeFileMode = 0750
	fileMode    = 0644
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

	log.Info().Str("basedir", baseDir).Msg("PrepareRclone")

	rcloneFile, errExePath := ExePath()
	if errExePath != nil {
		return errExePath
	}

	log.Info().Str("rcloneFile", rcloneFile).Msg("PrepareRclone")
	log.Info().Msg("PrepareRclone - get asset data")

	data, errAsset := rclone.Asset("data/darwin/rclone_osx")
	log.Info().Int("assetLen", len(data)).Msg("core darwin")

	if errAsset != nil {
		log.Err(errAsset).Msg("Rclone asset for Darwin not found")

		return fmt.Errorf("prepare rclone %w", errAsset)
	}

	log.Info().Msg("PrepareRclone - create executable")

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

	log.Info().Int64("writtenData", writtenData).Msg("core darwin")

	if errWrite != nil {
		log.Err(errWrite).Msg("Cannot write rclone executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errWrite)
	}

	rcloneExeFile.Close()

	log.Info().Msg("PrepareRclone - change executable mod")

	errChmod := os.Chmod(rcloneFile, os.FileMode(exeFileMode))
	if errChmod != nil {
		log.Err(errChmod).Msg("Cannot make rclone an executable in cache dir")

		return fmt.Errorf("prepare rclone %w", errChmod)
	}

	return nil
}

func MountVolume(instance string, remotePath string, localPath string, configPath string) error { // nolint: funlen
	log.Info().Msg("Prepare Rclone")
	errPrepare := PrepareRclone()
	if errPrepare != nil {
		log.Err(errPrepare).Msg("Cannot prepare Rclone")
		return errPrepare
	}

	log.Info().Msg("Get Rclone file path")
	rcloneFile, errExePath := ExePath()
	if errExePath != nil {
		return errExePath
	}

	log.Info().Msg("Make local dir")
	_, errLocalPath := os.Stat(localPath)
	if os.IsNotExist(errLocalPath) {
		errMkdir := os.MkdirAll(localPath, os.ModePerm)
		if errMkdir != nil {
			panic(errMkdir)
		}
	}

	conf := fmt.Sprintf("%s:%s", instance, remotePath)
	log.Info().Str("conf", conf).Msg("Prepare mounting points")

	configPathAbs, errConfigPath := filepath.Abs(path.Join(configPath, "/rclone.conf"))
	if errConfigPath != nil {
		log.Err(errConfigPath).Msg("server")
		return errConfigPath
	}

	logPath, errLogPath := filepath.Abs(path.Join(configPath, "/rclone.log"))
	if errLogPath != nil {
		log.Err(errLogPath).Msg("server")
		return errLogPath
	}

	localPathAbs, errLocalPath := filepath.Abs(localPath)
	if errLocalPath != nil {
		log.Err(errLocalPath).Msg("server")
		return errLocalPath
	}

	log.Info().Str("command", strings.Join([]string{
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
	}, " ")).Msg("Prepare rclone call")
	grepCmd := exec.Command(
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

	log.Info().Msg("Call rclone")

	errStart := grepCmd.Start()
	if errStart != nil {
		panic(errStart)
	}

	return nil
}
