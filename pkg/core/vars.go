package core

import _ "embed"

type InstanceInfo struct {
	Name     string
	LogFile  string
	Port     int
	Password bool
}

const (
	exeFileMode  = 0750
	fileMode     = 0644
	divider      = "------------------------------------------------------------------------------"
	logMaxSizeMB = 100
	oneMB        = 1000000
)

var (
	instanceLogFilename = "sts-wire.log"
	GitCommit           string //nolint:gochecknoglobals
	StsVersion          string //nolint:gochecknoglobals
	BuiltTime           string //nolint:gochecknoglobals
	OsArch              string //nolint:gochecknoglobals
	RcloneVersion       string //nolint:gochecknoglobals
)

// DATA
var (
	//go:embed "data/html/errorNoToken.html"
	htmlErrorNoToken []byte
	//go:embed "data/html/errorTokenExpired.html"
	htmlErrorTokenExpired []byte
	//go:embed "data/html/errorNoSaveToken.html"
	htmlErrorNoSaveToken []byte
	//go:embed "data/html/errorNoCred.html"
	htmlErrorNoCred []byte
	//go:embed "data/html/mountingPage.html"
	htmlMountingPage []byte
	//go:embed "data/html/errorNoStsCred.html"
	htmlErrorNoStsCred []byte
)
