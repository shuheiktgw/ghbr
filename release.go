package main

import (
	"github.com/spf13/cobra"
)

type releaseOptions struct {
	token, org, owner, repo, branch string
	force, merge                    bool
}

var releaseOpts releaseOptions

func NewReleaseCmd(generator GhbrGenerator) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "release",
		Aliases: []string{"update", "bumpup"},
		Short:   "Update your Homebrew formula to point to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRelease(generator)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	setReleasePreRunE(cmd)

	return cmd
}

func setReleasePreRunE(cmd *cobra.Command) {
	setReleaseFlags(cmd)
	err := validateReleaseFlags()

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err != nil {
			return cmdError{error: err, exitCode: ExitCodeParseFlagsError}
		}

		return err
	}
}

func runRelease(generator GhbrGenerator) error {
	g := generator(releaseOpts.token)

	lr, err := g.GetLatestRelease(releaseOpts.owner, releaseOpts.repo)
	if err != nil {
		return err
	}

	return g.UpdateFormula(releaseOpts.org, releaseOpts.owner, releaseOpts.repo, releaseOpts.branch, releaseOpts.force, releaseOpts.merge, lr)
}

func setReleaseFlags(cmd *cobra.Command) {
	// Set token flag
	setTokenFlag(cmd, &releaseOpts.token)

	// Set org flag
	cmd.Flags().StringVarP(&createOpts.org, "org", "g", "", "GitHub organization hosting a formula on")

	// Set owner flag
	setOwnerFlag(cmd, &releaseOpts.owner)

	// Set repository flag
	setRepositoryFlag(cmd, &releaseOpts.repo)

	// Set branch flag
	cmd.Flags().StringVarP(&releaseOpts.branch, "branch", "b", "master", "GitHub branch")

	// Set force flag
	cmd.Flags().BoolVarP(&releaseOpts.force, "force", "f", false, "Forcefully update a formula file, even if it's up-to-date")

	// Set merge flag
	cmd.Flags().BoolVarP(&releaseOpts.merge, "merge", "m", false, "Merge a Pull Request or not")
}

func validateReleaseFlags() error {
	// Token
	if err := validateToken(releaseOpts.token); err != nil {
		return err
	}

	// Owner
	if err := validateOwner(releaseOpts.owner); err != nil {
		return err
	}

	// Repository
	if err := validateRepository(releaseOpts.repo); err != nil {
		return err
	}

	return nil
}
