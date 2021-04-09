package main

import (
	"github.com/sumologic/sumo-cli/cmd"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {

	// Let's get going
	cmd.NewRootCmd(version, commit, date, builtBy).Execute()
}
