package cmd

import (
	"github.com/sumologic/sumo-cli/util"
	"testing"
)

func Test_PingSuccessful(t *testing.T) {
	util.OutputMatchesRegex(t, NewRootCmd("", "", "", ""),
		`^Pong! \(test https://.*?sumologic.*?/\)\n$`,
		"-p", "test", "ping")
}
