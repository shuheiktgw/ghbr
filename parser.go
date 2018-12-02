package main

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/spf13/cobra"
	"github.com/tcnksm/go-gitconfig"
)

var ownerNameRegex = regexp.MustCompile(`([-a-zA-Z0-9]+)/[^/]+$`)

func setTokenFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "token", "t", defaultToken(), "GitHub personal access token")
}

func setOwnerFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "owner", "o", defaultOwner(), "GitHub repository owner name")
}

func setRepositoryFlag(cmd *cobra.Command, dest *string) {
	cmd.Flags().StringVarP(dest, "repository", "r", defaultRepo(), "GitHub repository")
}

func validateToken(token string) error {
	if len(token) == 0 {
		return fmt.Errorf("missing GitHub personal access token\n\n"+
			"Please set it via `-t` option, %s environment variable or github.token in .gitconfig\n", EnvGitHubToken)
	}

	return nil
}

func validateOwner(owner string) error {
	if len(owner) == 0 {
		return errors.New("missing GitHub Repository Owner \n\n" +
			"Please set it via `-u` option.\n" +
			"You can set a default owner name in `github.username` or `user.name` in `~/.gitconfig` file\n")
	}

	return nil
}

func validateRepository(repo string) error {
	if len(repo) == 0 {
		return errors.New("missing GitHub repository\n\n" +
			"ghbr extracts repository from .git/config, so move to the root of your project," +
			"or set it via `-r` option")
	}

	return nil
}

func defaultToken() string {
	// First search for GITHUB_TOKEN environment variable
	t := os.Getenv(EnvGitHubToken)

	// Next search for github.token in .gitconfig
	if len(t) == 0 {
		t, _ = gitconfig.GithubToken()
	}

	return t
}

func defaultOwner() string {
	var owner string

	origin, err := gitconfig.OriginURL()

	if err == nil {
		owner = retrieveOwnerName(origin)
	}

	if len(owner) == 0 {
		owner, err = gitconfig.GithubUser()
		if err != nil {
			owner, err = gitconfig.Username()
		}
	}

	return owner
}

func retrieveOwnerName(repoURL string) string {
	matched := ownerNameRegex.FindStringSubmatch(repoURL)
	if len(matched) < 2 {
		return ""
	}
	return matched[1]
}

func defaultRepo() string {
	dr, _ := gitconfig.Repository()

	return dr
}
