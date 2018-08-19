package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shuheiktgw/ghbr/hbr"
)

const (
	testRepo   = "testRepo"
	testBranch = "master"
)

func TestRelease(t *testing.T) {
	cases := []struct {
		arg string

		expectedGhbrCount    int
		expectedErrorMessage string
	}{
		{arg: fmt.Sprintf("ghbr release -t test -u TestUser -r %s", testRepo), expectedGhbrCount: 1, expectedErrorMessage: ""},
		{arg: fmt.Sprintf("ghbr release -t test -r %s", testRepo), expectedGhbrCount: 1, expectedErrorMessage: ""},
	}

	for i, tc := range cases {
		generator, ctl := generateMockGHBR(t, tc.expectedGhbrCount)
		cmd := NewReleaseCmd(generator)

		args := strings.Split(tc.arg, " ")
		cmd.SetArgs(args[1:])

		if err := cmd.Execute(); err != nil {
			if want, got := tc.expectedErrorMessage, err.Error(); !strings.Contains(got, want) {
				t.Fatalf("#%d unexpected error occured: want: %s, got: %s", i, want, got)
			}
		}

		ctl.Finish()
	}
}

func generateMockGHBR(t *testing.T, count int) (hbr.Generator, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)

	return func(token, owner string) hbr.HBRWrapper {
		mockWrapper := hbr.NewMockHBRWrapper(mockCtrl)

		release := &hbr.LatestRelease{}
		mockWrapper.EXPECT().GetCurrentRelease(testRepo).Return(release).Times(count)
		mockWrapper.EXPECT().UpdateFormula(testRepo, testBranch, false, release).Times(count)
		mockWrapper.EXPECT().Err().Return(nil).Times(count)

		return mockWrapper
	}, mockCtrl
}
