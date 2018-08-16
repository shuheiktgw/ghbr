package cmd

import (
	"testing"
	"bytes"
	"strings"
	"github.com/shuheiktgw/ghbr/version"
)

func TestVersion (t *testing.T) {
	cmd := NewVersionCmd()

	buf := new(bytes.Buffer)
	cmd.SetOutput(buf)

	args := strings.Split("ghbr version", " ")
	cmd.SetArgs(args[1:])

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error occured: %s", err)
	}

	if got, want := buf.String(), version.OutputVersion(); got != want {
		t.Fatalf("invalid response: got: %s, want: %s", got, want)
	}
}


