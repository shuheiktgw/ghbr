package cmd

import (
	"github.com/shuheiktgw/ghbr/version"

	"github.com/spf13/cobra"
	)

func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the current version of GHBR",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Print(version.OutputVersion())
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return cmd
}
