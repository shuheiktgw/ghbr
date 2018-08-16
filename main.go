package main

import (
	"os"
	"github.com/shuheiktgw/ghbr/cmd"
)

const EnvGitHubToken = "GITHUB_TOKEN"

func main() {
	os.Exit(cmd.Execute())
}
