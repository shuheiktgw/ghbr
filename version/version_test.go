package version

import (
	"fmt"
	"testing"
)

func TestVersion_OutputVersion(t *testing.T) {
	if got, want := OutputVersion(), fmt.Sprintf("The current version of %s is v%s\n", Name, Version); got != want {
		t.Fatalf("Unexpected version string: want: %s, got: %s", want, got)
	}
}
