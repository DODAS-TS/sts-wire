package main

import (
	"github.com/DODAS-TS/sts-wire/pkg/core"
	_ "github.com/DODAS-TS/sts-wire/pkg/core"
	"github.com/awnumar/memguard"
	_ "github.com/go-bindata/go-bindata"
)

func main() {
	// Safely terminate in case of an interrupt signal
	memguard.CatchInterrupt()

	// Purge the session when we return
	defer memguard.Purge()

	core.Execute()
}
