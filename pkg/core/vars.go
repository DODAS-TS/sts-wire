package core

const (
	exeFileMode = 0750
	fileMode    = 0644
	divider     = "------------------------------------------------------------------------------"
)

var (
	instanceLogFilename = ".instance.log"
	GitCommit           string //nolint:gochecknoglobals
	StsVersion          string //nolint:gochecknoglobals
	BuiltTime           string //nolint:gochecknoglobals
	OsArch              string //nolint:gochecknoglobals
	RcloneVersion       string //nolint:gochecknoglobals
)
