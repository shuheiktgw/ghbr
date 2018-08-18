package cmd

import (
	"os"
	"fmt"

	"github.com/shuheiktgw/ghbr/hbr"

	"github.com/spf13/cobra"
	"github.com/tcnksm/go-gitconfig"
	"github.com/pkg/errors"
)

type releaseOptions struct {
	token, owner, repo, branch string
}

var options releaseOptions

func NewReleaseCmd(generator hbr.Generator) *cobra.Command {
	cmd := &cobra.Command{
		Use: "release",
		Aliases: []string{"update", "bumpup"},
		Short: "Update your Homebrew formula to point to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelease(generator)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	setPreRunE(cmd)

	return cmd
}

func setPreRunE(cmd *cobra.Command) {
	err := parseFlags(cmd)

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err != nil {
			return cmdError{error: err, exitCode: ExitCodeParseFlagsError}
		}

		return err
	}
}

func runRelease(generator hbr.Generator) error {
	g := generator(options.token, options.owner)
	lr := g.GetCurrentRelease(options.repo)
	g.UpdateFormula(options.repo, options.branch, lr)

	if err := g.Err(); err != nil {
		return err
	}

	return nil
}

func parseFlags(cmd *cobra.Command) error{
	// Token
	dt, err := defaultToken()

	if err != nil {
		return err
	}

	cmd.Flags().StringVarP(&options.token, "token", "t", dt, "GitHub personal access token")

	if len(options.token) == 0 {
		return fmt.Errorf("missing GitHub personal access token\n\n" +
			"Please set it via `-t` option, %s environment variable or github.token in .gitconfig\n", EnvGitHubToken)
	}

	// Owner
	do, err := gitconfig.Username()

	if err != nil {
		return err
	}

	cmd.Flags().StringVarP(&options.owner, "username", "u", do, "GitHub username")

	if len(options.owner) == 0 {
		return errors.New("missing GitHub username\n\n" +
			"ghbr extracts username from .git/config, so move to the root of your project," +
			"or set it via `-u` option")
	}

	// Repository
	dr, err := gitconfig.Repository()

	if err != nil {
		return err
	}

	cmd.Flags().StringVarP(&options.repo, "repository", "r", dr, "GitHub repository")

	if len(options.repo) == 0 {
		return errors.New("missing GitHub repository\n\n" +
			"ghbr extracts repository from .git/config, so move to the root of your project," +
			"or set it via `-r` option")
	}

	// Branch
	cmd.Flags().StringVarP(&options.branch, "branch", "b", "master", "GitHub branch")

	return nil
}

func defaultToken() (string, error) {
	// First search for GITHUB_TOKEN environment variable
	t := os.Getenv(EnvGitHubToken)

	// Next search for github.token in .gitconfig
	if len(t) == 0 {
		return gitconfig.GithubToken()
	}

	return t, nil
}