package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/DODAS-TS/sts-wire/pkg/core"
	_ "github.com/DODAS-TS/sts-wire/pkg/core"
	"github.com/awnumar/memguard"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			if !strings.Contains(fmt.Sprintf("%v", err), "unknown flag") {
				core.WriteReport(err)
				panic(err)
			}
		}
	}()

	// Safely terminate in case of an interrupt signal
	memguard.CatchSignal(func(_ os.Signal) {}, os.Interrupt)

	// Purge the session when we return
	defer memguard.Purge()

	core.Execute()
}
