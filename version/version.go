package version

import (
	"bytes"
	"fmt"
	"time"

	"github.com/tcnksm/go-latest"
)

// The current version of ghbr
const Version = "0.0.8"

// The owner of ghbr
const Owner = "shuheiktgw"

// Name is the name of this application
const Name = "ghbr"

// OutputVersion outputs current version wof ghbr. It also checks
// the latest release and adds a warning to update ghbr
func OutputVersion() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "The current version of %s is v%s\n", Name, Version)

	// Get the latest release
	verCheckCh := make(chan *latest.CheckResponse)
	go func() {
		githubTag := &latest.GithubTag{
			Owner:      Owner,
			Repository: Name,
		}

		res, err := latest.Check(githubTag, Version)

		// Ignore the error
		if err != nil {
			return
		}

		verCheckCh <- res
	}()

	select {
	case <-time.After(2 * time.Second):
	case res := <-verCheckCh:
		if res.Outdated {
			fmt.Fprintf(&b, "The Latest version is v%s, please update ghbr\n", res.Current)
		}
	}

	return b.String()
}
