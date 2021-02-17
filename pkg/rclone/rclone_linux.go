package rclone

import _ "embed"

//go:embed "data/linux/rclone"
var Executable []byte
