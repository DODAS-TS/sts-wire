package rclone

import _ "embed"

//go:embed "data/windows/rclone"
var Executable []byte
