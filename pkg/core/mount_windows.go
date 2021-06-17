// +build windows
package core

const (
	rcloneExeString = "rclone.exe"
)

func checkMountpoint(path string) (bool, error) {
	log.Warn().Msg("checkMountpoint not yet implemented on windows...")

	return true, nil
}

func unmount() error {
	log.Warn().Msg("unmount not yet implemented on windows...")

	return nil
}
