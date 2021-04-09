package util

import (
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"regexp"
	"testing"
)

func OutputMatchesRegex(t *testing.T, cmd *cobra.Command, pattern string, args ...string) {
	out := ExecuteCommand(t, cmd, args)
	matched, err := regexp.MatchString(pattern, out)
	if err != nil {
		t.Errorf("Regex error: %v", err)
	}
	if !matched {
		t.Errorf("Output\n%s\ndid not match\n%s\n", pattern, out)
	}
}

func ExecuteCommand(t *testing.T, root *cobra.Command, args []string) (output string) {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	root.SetArgs(args)
	err := root.Execute()
	if err != nil {
		t.Errorf("Cobra error: %v", err)
	}

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	outString := string(out)
	return outString
}
