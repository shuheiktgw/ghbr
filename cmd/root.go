package cmd

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ExitCodeOK    = iota
	ExitCodeError
)

var cfgFile string

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
	cobra.OnInitialize(initConfig)

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ghbr.yaml)")

	cmd.AddCommand(NewVersionCmd())
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

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".ghbr")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
