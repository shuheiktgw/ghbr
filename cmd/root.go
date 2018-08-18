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

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ghbr",
		Short: "GHBR is a simple CLI tool to create and update your Homebrew formula",
	}

	cmd.AddCommand(NewVersionCmd())
	cmd.AddCommand(NewReleaseCmd(hbr.GenerateGHBR))
	return cmd
}

func Execute() int {
	cmd := NewRootCmd()
	cmd.SetOutput(os.Stdout)
	if err := cmd.Execute(); err != nil {
		cmd.SetOutput(os.Stderr)
		cmd.Println(err)

		if e, ok := err.(cmdError); ok {
			return e.exitCode
		}

		return ExitCodeError
	}

	return ExitCodeOK
}
