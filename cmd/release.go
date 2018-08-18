package cmd

import (
	"os"
	"fmt"

	"github.com/shuheiktgw/ghbr/ghbr"

	"github.com/spf13/cobra"
	"github.com/tcnksm/go-gitconfig"
	"github.com/pkg/errors"
)

type releaseOptions struct {
	token, owner, repo, branch string
}

func NewReleaseCmd(generator ghbr.Generator) *cobra.Command {
	cmd := &cobra.Command{
		Use: "release",
		Aliases: []string{"update", "bumpup"},
		Short: "Update your Homebrew formula to point to the latest release",
		Run: func(cmd *cobra.Command, args []string) {
			runRelease(cmd, args, generator)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}

func runRelease(cmd *cobra.Command, _ []string, generator ghbr.Generator) {
	options, err := parseFlags(cmd)

	if err != nil {
		cmd.Print(err)
		return
	}

	g := generator(options.token, options.owner)
	lr := g.GetCurrentRelease(options.repo)
	g.UpdateFormula(options.repo, options.branch, lr)

	if err := g.Err(); err != nil {
		cmd.Print(err)
	}

	return
}

func parseFlags(cmd *cobra.Command) (*releaseOptions, error) {
	var options releaseOptions

	// Token
	dt, err := defaultToken()

	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&options.token, "token", "t", dt, "GitHub personal access token")

	if len(options.token) == 0 {
		return nil, fmt.Errorf("missing GitHub personal access token\n\n" +
			"Please set it via `-t` option, %s environment variable or github.token in .gitconfig\n", EnvGitHubToken)
	}

	// Owner
	do, err := gitconfig.Username()

	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&options.owner, "username", "u", do, "GitHub username")

	if len(options.owner) == 0 {
		return nil, errors.New("missing GitHub username\n\n" +
			"ghbr extracts username from .git/config, so move to the root of your project," +
			"or set it via `-u` option")
	}

	// Repository
	dr, err := gitconfig.Repository()

	if err != nil {
		return nil, err
	}

	cmd.Flags().StringVarP(&options.repo, "repository", "r", dr, "GitHub repository")

	if len(options.repo) == 0 {
		return nil, errors.New("missing GitHub repository\n\n" +
			"ghbr extracts repository from .git/config, so move to the root of your project," +
			"or set it via `-r` option")
	}

	// Branch
	cmd.Flags().StringVarP(&options.branch, "branch", "b", "master", "GitHub branch")

	return &options, nil
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