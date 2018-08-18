package cmd

import (
	"testing"
	"strings"
	"github.com/shuheiktgw/ghbr/hbr"
	"github.com/golang/mock/gomock"
	"fmt"
	)

const (
	testRepo   = "testRepo"
	testBranch = "master"
)

func TestRelease(t *testing.T) {
	cases := []struct {
		arg       string
		ghbrError error

		expectedGhbrCount    int
		expectedErrorMessage string
	}{
		{arg: fmt.Sprintf("ghbr release -t test -u TestUser -r %s", testRepo), ghbrError: nil, expectedGhbrCount: 1, expectedErrorMessage: ""},
	}

	for i, tc := range cases {
		generator, ctl := generateMockGHBR(t, tc.expectedGhbrCount, tc.ghbrError)
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

func generateMockGHBR(t *testing.T, count int, err error) (hbr.Generator, *gomock.Controller) {
	mockCtrl := gomock.NewController(t)

	return func(token, owner string) hbr.GHBRWrapper {
		mockWrapper := hbr.NewMockGHBRWrapper(mockCtrl)

		release := &hbr.LatestRelease{}
		mockWrapper.EXPECT().GetCurrentRelease(testRepo).Return(release).Times(count)
		mockWrapper.EXPECT().UpdateFormula(testRepo, testBranch, release).Times(count)
		mockWrapper.EXPECT().Err().Return(err).Times(count)

		return mockWrapper
	}, mockCtrl
}
