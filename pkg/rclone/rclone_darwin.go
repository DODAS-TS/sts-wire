package rclone

import _ "embed"

//go:embed "data/darwin/rclone"
var Executable []byte
