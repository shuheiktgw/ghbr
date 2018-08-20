package cmd

import (
	"github.com/shuheiktgw/ghbr/hbr"
	"github.com/spf13/cobra"
)

type createOptions struct {
	token, owner, repo, font string
	private                  bool
}

var createOpts createOptions

func NewCreateCmd(generator hbr.Generator) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"init"},
		Short:   "Create a GitHub repository to host a Homebrew formula",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreate(generator)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	setCreatePreRunE(cmd)

	return cmd
}

func setCreatePreRunE(cmd *cobra.Command) {
	setCreateFlags(cmd)
	err := validateCreateFlags()

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err != nil {
			return cmdError{error: err, exitCode: ExitCodeParseFlagsError}
		}

		return err
	}
}

func runCreate(generator hbr.Generator) error {
	g := generator(createOpts.token, createOpts.owner)
	lr := g.GetCurrentRelease(createOpts.repo)
	g.CreateFormula(createOpts.repo, createOpts.font, createOpts.private, lr)

	if err := g.Err(); err != nil {
		return err
	}

	return nil
}

func setCreateFlags(cmd *cobra.Command) {
	// Set token flag
	setTokenFlag(cmd, &createOpts.token)

	// Set owner flag
	setOwnerFlag(cmd, &createOpts.owner)

	// Set repository flag
	setRepositoryFlag(cmd, &createOpts.repo)

	// Ascii Font
	cmd.Flags().StringVarP(&createOpts.font, "font", "f", "isometric3", "caveats Ascii Font from go-figure")

	// Repository private setting
	cmd.Flags().BoolVarP(&createOpts.private, "private", "p", false, "If true, GHBR creates a private repository on GitHub")
}

func validateCreateFlags() error {
	// Token
	if err := validateToken(createOpts.token); err != nil {
		return err
	}

	// Owner
	if err := validateOwner(createOpts.owner); err != nil {
		return err
	}

	// Repository
	if err := validateRepository(createOpts.repo); err != nil {
		return err
	}

	return nil
}
