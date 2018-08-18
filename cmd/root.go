package cmd

import (
		"os"

		"github.com/spf13/cobra"
	"github.com/shuheiktgw/ghbr/hbr"
)

const (
	ExitCodeOK int = 0

	// Error Starts from 10
	ExitCodeError = 10 + iota
	ExitCodeParseFlagsError
)

const EnvGitHubToken = "GITHUB_TOKEN"

type cmdError struct {
	error
	exitCode int
}

func (cr cmdError) Error() string {
	return cr.Error()
}

var RootCmd = &cobra.Command{
	Use:   "ghbr",
	Short: "GHBR is a simple CLI tool to create and update your Homebrew formula",
}

func init() {
	RootCmd.AddCommand(NewVersionCmd())
	RootCmd.AddCommand(NewReleaseCmd(hbr.GenerateGHBR))
}

func Execute() int {
	RootCmd.SetOutput(os.Stdout)
	if err := RootCmd.Execute(); err != nil {
		RootCmd.SetOutput(os.Stderr)
		RootCmd.Println(err)

		if e, ok := err.(cmdError); ok {
			return e.exitCode
		}

		return ExitCodeError
	}

	return ExitCodeOK
}
