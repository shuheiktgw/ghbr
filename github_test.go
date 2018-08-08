package main

import (
	"testing"
	"os"
	)

const (
	TestOwner = "shuheiktgwtest"
	TestLibraryRepo = "github-api-test-go-homebrew"
	TestFormulaRepo = "homebrew-" + TestLibraryRepo
)

func testGitHubClient(t *testing.T) *GitHubClient {
	token := os.Getenv(EnvGitHubToken)
	client, err := NewGitHubClient(TestOwner, token)
	if err != nil {
		t.Fatal("NewGitHubClient failed:", err)
	}
	return client
}

func TestNewGitHubClientFail(t *testing.T) {
	cases := []struct {
		owner, token string
	}{
		{owner: "", token: "testToken"},
		{owner: "testOwner", token: ""},
	}

	for i, tc := range cases {
		if _, err := NewGitHubClient(tc.owner, tc.token); err == nil {
			t.Fatalf("#%d NewGitHubClient: error is not supposed to be nil", i)
		}
	}
}

func TestNewGitHubClientSuccess(t *testing.T) {
	if _, err := NewGitHubClient("testOwner", "testToken"); err != nil {
		t.Fatalf("NewGitHubClient: unexpected error occurred: %s", err)
	}
}

func TestGetLatestReleaseFail(t *testing.T) {
	cases := []struct {
		repo string
	}{
		{repo: "unknowwn"},
		{repo: ""},
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if _, err := c.GetLatestRelease(tc.repo); err == nil {
			t.Fatalf("#%d GetLatestRelease: error is not supposed to be nil", i)
		}
	}
}

func TestGetLatestReleaseSuccess(t *testing.T) {
	c := testGitHubClient(t)

	if _, err := c.GetLatestRelease(TestLibraryRepo); err != nil {
		t.Fatalf("GetLatestRelease: unexpected error occured: %s", err)
	}
}

func TestCreateBranchFail(t *testing.T) {
	cases := []struct {
		repo, origin, new string
	}{
		{repo: "unknowwn", origin: "master", new: "test" },
		{repo: "", origin: "master", new: "test" },
		{repo: TestLibraryRepo, origin: "unknown", new: "test" },
		{repo: TestLibraryRepo, origin: "", new: "test" },
		{repo: TestLibraryRepo, origin: "master", new: "" },
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if err := c.CreateBranch(tc.repo, tc.origin, tc.new); err == nil {
			if e := c.DeleteLatestRef(tc.repo, tc.new); e != nil {
				t.Errorf("#%d DeleteLatestRef: failed to rollback CreateBranch: %s", i, e)
			}
			t.Fatalf("#%d CreateBranch: error is not supposed to be nil", i)
		}
	}
}

func TestDeleteLatestRefFail(t *testing.T) {
	cases := []struct {
		repo, branch string
	}{
		{repo: "unknowwn", branch: "master"},
		{repo: "", branch: "master"},
		{repo: TestLibraryRepo, branch: "unknown"},
		{repo: TestLibraryRepo, branch: ""},
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if err := c.DeleteLatestRef(tc.repo, tc.branch); err == nil {
			t.Fatalf("#%d DeleteLatestRef: error is not supposed to be nil", i)
		}
	}
}

func TestCreateAndDeleteBranch(t *testing.T) {
	c := testGitHubClient(t)

	err := c.CreateBranch(TestLibraryRepo, "master", "test")

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	err = c.DeleteLatestRef(TestLibraryRepo, "test")

	if err != nil {
		t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
	}
}




