package main

import (
	"github/GAtom22/missedblocks/app"
	"github/GAtom22/missedblocks/cli"
	"os"
	"strings"
)

// Possible app modes are CLI or standalone
// if APP_MODE env variable not provided
// defaults to standalone
const cliMode = "cli"

func main() {
	mode := os.Getenv("APP_MODE")
	if strings.ToLower(mode) == cliMode {
		cli.Run()
	} else {
		configPath := os.Getenv("CONFIG_PATH")
		app.Run(configPath)
	}
}
