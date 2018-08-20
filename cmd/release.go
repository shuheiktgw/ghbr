package cmd

import (
	"github.com/shuheiktgw/ghbr/hbr"
	"github.com/spf13/cobra"
)

type releaseOptions struct {
	token, owner, repo, branch string
	merge                      bool
}

var releaseOpts releaseOptions

func NewReleaseCmd(generator hbr.Generator) *cobra.Command {
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

func runRelease(generator hbr.Generator) error {
	g := generator(releaseOpts.token, releaseOpts.owner)
	lr := g.GetCurrentRelease(releaseOpts.repo)
	g.UpdateFormula(releaseOpts.repo, releaseOpts.branch, releaseOpts.merge, lr)

	if err := g.Err(); err != nil {
		return err
	}

	return nil
}

func setReleaseFlags(cmd *cobra.Command) {
	// Set token flag
	setTokenFlag(cmd, &releaseOpts.token)

	// Set owner flag
	setOwnerFlag(cmd, &releaseOpts.owner)

	// Set repository flag
	setRepositoryFlag(cmd, &releaseOpts.repo)

	// Set branch flag
	cmd.Flags().StringVarP(&releaseOpts.branch, "branch", "b", "master", "GitHub branch")

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
