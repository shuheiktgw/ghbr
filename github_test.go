package main

import (
	"testing"
	"os"
	"encoding/base64"
	"time"
)

const (
	TestOwner = "shuheiktgwtest"
	TestRepo  = "github-api-test-go-homebrew"
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

	if _, err := c.GetLatestRelease(TestRepo); err != nil {
		t.Fatalf("GetLatestRelease: unexpected error occured: %s", err)
	}
}

func TestCreateBranchFail(t *testing.T) {
	cases := []struct {
		repo, origin, new string
	}{
		{repo: "unknowwn", origin: "master", new: "test" },
		{repo: "", origin: "master", new: "test" },
		{repo: TestRepo, origin: "unknown", new: "test" },
		{repo: TestRepo, origin: "", new: "test" },
		{repo: TestRepo, origin: "master", new: "" },
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
		{repo: TestRepo, branch: "unknown"},
		{repo: TestRepo, branch: ""},
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

	err := c.CreateBranch(TestRepo, "master", "test")

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	err = c.DeleteLatestRef(TestRepo, "test")

	if err != nil {
		t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
	}
}

func TestCreatePullRequestFail(t *testing.T) {
	cases := []struct {
		repo, title, head, base, body string
	}{
		{repo: "unknowwn", title: "Test PR!", head: "develop", base: "master", body: "This is a test PR!"},
		{repo: "", title: "Test PR!", head: "develop", base: "master", body: "This is a test PR!"},
		{repo: TestRepo, title: "", head: "develop", base: "master", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "unknown", base: "master", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "", base: "master", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "develop", base: "unknown", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "develop", base: "", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "develop", base: "master", body: ""},
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if pr, err := c.CreatePullRequest(tc.repo, tc.title, tc.head, tc.base, tc.body); err == nil {
			if e := c.ClosePullRequest(tc.repo, *pr.Number); e != nil {
				t.Errorf("#%d ClosePullRequest: failed to rollback CreatePullRequest: %s", i, e)
			}
			t.Fatalf("#%d CreatePullRequest: error is not supposed to be nil", i)
		}
	}
}

func TestMergePullRequestFail(t *testing.T) {
	cases := []struct {
		repo string
		number int
	}{
		{repo: "unknowwn", number: 100},
		{repo: "", number: 100},
		{repo: TestRepo, number: 100},
		{repo: TestRepo, number: 0},
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if err := c.MergePullRequest(tc.repo, tc.number); err == nil {
			t.Fatalf("#%d MergePullRequest: error is not supposed to be nil", i)
		}
	}
}

func TestClosePullRequestFail(t *testing.T) {
	cases := []struct {
		repo string
		number int
	}{
		{repo: "unknowwn", number: 100},
		{repo: "", number: 100},
		{repo: TestRepo, number: 100},
		{repo: TestRepo, number: 0},
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if err := c.ClosePullRequest(tc.repo, tc.number); err == nil {
			t.Fatalf("#%d ClosePullRequest: error is not supposed to be nil", i)
		}
	}
}

func TestCreateAndMergeAndClosePullRequestSuccess(t *testing.T) {
	c := testGitHubClient(t)

	masterReplica, developReplica := "master_replica", "develop_replica"

	// Create new branches for this test
	err := c.CreateBranch(TestRepo, "master", masterReplica)

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	err = c.CreateBranch(TestRepo, "develop", developReplica)

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	// Create PR develop_replica -> master_replica
	developRepToMasterRepPR, err := c.CreatePullRequest(TestRepo, "First Test PR for TestCreateAndMergeAndClosePullRequest", developReplica, masterReplica, "Test PR!")

	if err != nil {
		t.Fatalf("CreatePullRequest: unexpected error occured: %s", err)
	}

	// Wait a few seconds to prevent 405 Error
	time.Sleep(3 * time.Second)

	// Merge PR develop_replica -> master_replica
	err = c.MergePullRequest(TestRepo, *developRepToMasterRepPR.Number)

	if err != nil {
		t.Fatalf("MergePullRequest: unexpected error occured: %s", err)
	}

	// Create PR master_replica -> master
	masterRepToMasterPR, err := c.CreatePullRequest(TestRepo, "Second Test PR for TestCreateAndMergeAndClosePullRequest", masterReplica, "master", "Test PR!")

	if err != nil {
		t.Fatalf("CreatePullRequest: unexpected error occured: %s", err)
	}

	// Close PR master_replica -> master
	err = c.ClosePullRequest(TestRepo, *masterRepToMasterPR.Number)

	if err != nil {
		t.Fatalf("MergePullRequest: unexpected error occured: %s", err)
	}

	// Clean up the branches created for this test
	err = c.DeleteLatestRef(TestRepo, masterReplica)

	if err != nil {
		t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
	}

	err = c.DeleteLatestRef(TestRepo, developReplica)

	if err != nil {
		t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
	}
}

func TestGetFileFail(t *testing.T) {
	cases := []struct {
		repo, branch, path string
	}{
		{repo: "unknowwn", branch: "master", path: "main.go"},
		{repo: "", branch: "master", path: "main.go"},
		{repo: TestRepo, branch: "unknown", path: "main.go"},
		{repo: TestRepo, branch: "", path: "main.go"},
		{repo: TestRepo, branch: "master", path: "unknown.go"},
		{repo: TestRepo, branch: "master", path: ""},
	}

	for i, tc := range cases {
		c := testGitHubClient(t)

		if _, err := c.GetFile(tc.repo, tc.branch, tc.path); err == nil {
			t.Fatalf("#%d GetFile: error is not supposed to be nil", i)
		}
	}
}

func TestGetFileSuccess(t *testing.T) {
	c := testGitHubClient(t)

	file, err := c.GetFile(TestRepo, "master", "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	expected := "main.go"
	if got := *file.Name; got != expected {
		t.Fatalf("GetFile: unexpected file received: expected: %s, got: %s", expected, got)
	}
}

func TestUpdateFileFail(t *testing.T) {
	c := testGitHubClient(t)
	f, err := c.GetFile(TestRepo, "master", "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	cases := []struct {
		repo, branch, path, sha, message, content string
	}{
		{repo: "unknowwn", branch: "master", path: "main.go", sha: *f.SHA, message: "test!", content: "test"},
		{repo: "", branch: "master", path: "main.go", sha: *f.SHA, message: "test!", content: "test"},
		{repo: TestRepo, branch: "unknown", path: "main.go", sha: *f.SHA, message: "test!", content: "test"},
		{repo: TestRepo, branch: "", path: "main.go", sha: *f.SHA, message: "test!", content: "test"},
		{repo: TestRepo, branch: "master", path: "", sha: *f.SHA, message: "test!", content: "test"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: "unknown", message: "test!", content: "test"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: "", message: "test!", content: "test"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: "unknown", message: "test!", content: "test"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: *f.SHA, message: "", content: "test"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: *f.SHA, message: "test!", content: ""},
	}

	for i, tc := range cases {
		if err := c.UpdateFile(tc.repo, tc.branch, tc.path, tc.sha, tc.message, []byte(tc.content)); err == nil {
			t.Fatalf("#%d UpdateFile: error is not supposed to be nil", i)
		}
	}
}

func TestUpdateFileSuccess(t *testing.T) {
	c := testGitHubClient(t)

	testBranch := "test_update_file"

	// Create new branches for this test
	err := c.CreateBranch(TestRepo, "master", testBranch)

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	// Delete the branches created for this test
	defer c.DeleteLatestRef(TestRepo, testBranch)

	// Get main.go on the test branch
	f, err := c.GetFile(TestRepo, testBranch, "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	// Update main.go on the test branch
	err = c.UpdateFile(TestRepo, testBranch, "main.go", *f.SHA, "Update main.go", []byte("test!"))

	if err != nil {
		t.Fatalf("UpdateFile: unexpected error occured: %s", err)
	}

	// Get updated main.go on the test branch
	f, err = c.GetFile(TestRepo, testBranch, "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	// Check if the content is as expected
	decoded, err := base64.StdEncoding.DecodeString(*f.Content)

	if err != nil {
		t.Fatalf("failed to decode main.go: %s", err)
	}

	if string(decoded) != "test!" {
		t.Fatalf("unexpected content: got: %s, want: %s", string(decoded), "test!")
	}
}




