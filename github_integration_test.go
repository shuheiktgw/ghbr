// +build integration

package main

import (
	"encoding/base64"
	"os"
	"testing"
	"time"
)

const (
	IntegrationTestOwner       = "shuheiktgwtest"
	IntegrationTestRepo        = "github-api-test-go-homebrew"
	IntegrationTestGitHubToken = "GITHUB_TOKEN"
)

func testGitHubClient() *GitHubClient {
	token := os.Getenv(IntegrationTestGitHubToken)
	return NewGitHubClient(token)
}

func TestGetLatestReleaseFail(t *testing.T) {
	c := testGitHubClient()

	if _, err := c.GetLatestRelease(IntegrationTestOwner, "unknown"); err == nil {
		t.Fatalf("#GetCurrentRelease did not return error")
	}
}

func TestGetLatestReleaseSuccess(t *testing.T) {
	c := testGitHubClient()

	if _, err := c.GetLatestRelease(IntegrationTestOwner, IntegrationTestRepo); err != nil {
		t.Fatalf("GetCurrentRelease: unexpected error occured: %s", err)
	}
}

func TestCreateBranchFail(t *testing.T) {
	cases := []struct {
		repo, origin, new string
	}{
		{repo: "unknowwn", origin: "master", new: "test"},
		{repo: IntegrationTestRepo, origin: "unknown", new: "test"},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if err := c.CreateBranch(IntegrationTestOwner, tc.repo, tc.origin, tc.new); err == nil {
			if e := c.DeleteLatestRef(IntegrationTestOwner, tc.repo, tc.new); e != nil {
				t.Errorf("#%d #DeleteLatestRef failed to rollback CreateBranch: %s", i, e)
			}
			t.Fatalf("#%d #CreateBranch did not return error", i)
		}
	}
}

func TestDeleteLatestRefFail(t *testing.T) {
	cases := []struct {
		repo, branch string
	}{
		{repo: "unknowwn", branch: "master"},
		{repo: IntegrationTestRepo, branch: "unknown"},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if err := c.DeleteLatestRef(IntegrationTestOwner, tc.repo, tc.branch); err == nil {
			t.Fatalf("#%d #DeleteLatestRef did not return error", i)
		}
	}
}

func TestCreateAndDeleteBranch(t *testing.T) {
	c := testGitHubClient()

	err := c.CreateBranch(IntegrationTestOwner, IntegrationTestRepo, "master", "test")

	if err != nil {
		t.Fatalf("#CreateBranch returns unexpected error: %s", err)
	}

	err = c.DeleteLatestRef(IntegrationTestOwner, IntegrationTestRepo, "test")

	if err != nil {
		t.Fatalf("#DeleteLatestRef returns unexpected error: %s", err)
	}
}

func TestCreatePullRequestFail(t *testing.T) {
	cases := []struct {
		repo, title, head, base, body string
	}{
		{repo: "unknowwn", title: "Test PR!", head: "develop", base: "master", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "unknown", base: "master", body: "This is a test PR!"},
		{repo: TestRepo, title: "Test PR!", head: "develop", base: "unknown", body: "This is a test PR!"},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if pr, err := c.CreatePullRequest(IntegrationTestOwner, tc.repo, tc.title, tc.head, tc.base, tc.body); err == nil {
			if e := c.ClosePullRequest(IntegrationTestOwner, tc.repo, *pr.Number); e != nil {
				t.Errorf("#%d #ClosePullRequest failed to rollback #CreatePullRequest: %s", i, e)
			}
			t.Fatalf("#%d #CreatePullRequest did not return error", i)
		}
	}
}

func TestMergePullRequestFail(t *testing.T) {
	cases := []struct {
		repo   string
		number int
	}{
		{repo: "unknowwn", number: 99999},
		{repo: TestRepo, number: 99999},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if err := c.MergePullRequest(IntegrationTestRepo, tc.repo, tc.number); err == nil {
			t.Fatalf("#%d #MergePullRequest did not reutrn error", i)
		}
	}
}

func TestClosePullRequestFail(t *testing.T) {
	cases := []struct {
		repo   string
		number int
	}{
		{repo: "unknowwn", number: 9999},
		{repo: TestRepo, number: 99999},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if err := c.ClosePullRequest(IntegrationTestOwner, tc.repo, tc.number); err == nil {
			t.Fatalf("#%d #ClosePullRequest did not return error", i)
		}
	}
}

func TestCreateAndMergeAndClosePullRequestSuccess(t *testing.T) {
	c := testGitHubClient()

	masterReplica, developReplica := "master_replica", "develop_replica"

	// Create new branches for this test
	err := c.CreateBranch(IntegrationTestOwner, IntegrationTestRepo, "master", masterReplica)

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	err = c.CreateBranch(IntegrationTestOwner, IntegrationTestRepo, "develop", developReplica)

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	// Clean up the branches created for this test
	defer func() {
		err = c.DeleteLatestRef(IntegrationTestOwner, IntegrationTestRepo, masterReplica)

		if err != nil {
			t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
		}

		err = c.DeleteLatestRef(IntegrationTestOwner, IntegrationTestRepo, developReplica)

		if err != nil {
			t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
		}
	}()

	// Create PR develop_replica -> master_replica
	developRepToMasterRepPR, err := c.CreatePullRequest(IntegrationTestOwner, IntegrationTestRepo, "First Test PR for TestCreateAndMergeAndClosePullRequest", developReplica, masterReplica, "Test PR!")

	if err != nil {
		t.Fatalf("CreatePullRequest: unexpected error occured: %s", err)
	}

	// Merge PR develop_replica -> master_replica
	err = c.MergePullRequest(IntegrationTestOwner, IntegrationTestRepo, *developRepToMasterRepPR.Number)

	if err != nil {
		t.Fatalf("MergePullRequest: unexpected error occured: %s", err)
	}

	// Create PR master_replica -> master
	masterRepToMasterPR, err := c.CreatePullRequest(IntegrationTestOwner, IntegrationTestRepo, "Second Test PR for TestCreateAndMergeAndClosePullRequest", masterReplica, "master", "Test PR!")

	if err != nil {
		t.Fatalf("CreatePullRequest: unexpected error occured: %s", err)
	}

	// Close PR master_replica -> master
	err = c.ClosePullRequest(IntegrationTestOwner, IntegrationTestRepo, *masterRepToMasterPR.Number)

	if err != nil {
		t.Fatalf("MergePullRequest: unexpected error occured: %s", err)
	}
}

func TestGetFileFail(t *testing.T) {
	cases := []struct {
		repo, branch, path string
	}{
		{repo: "unknowwn", branch: "master", path: "main.go"},
		{repo: TestRepo, branch: "unknown", path: "main.go"},
		{repo: TestRepo, branch: "master", path: "unknown.go"},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if _, err := c.GetFile(IntegrationTestOwner, tc.repo, tc.branch, tc.path); err == nil {
			t.Fatalf("#%d GetFile: error is not supposed to be nil", i)
		}
	}
}

func TestGetFileSuccess(t *testing.T) {
	c := testGitHubClient()

	file, err := c.GetFile(IntegrationTestOwner, IntegrationTestRepo, "master", "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	expected := "main.go"
	if got := *file.Name; got != expected {
		t.Fatalf("GetFile: unexpected file received: expected: %s, got: %s", expected, got)
	}
}

func TestGitHubClient_CreateFileFail(t *testing.T) {
	cases := []struct {
		repo, branch, path, message, content string
	}{
		{repo: "unknown", branch: "master", path: "main.go", message: "test!", content: "test"},
		{repo: TestRepo, branch: "unknown", path: "main.go", message: "test!", content: "test"},
	}

	for i, tc := range cases {
		c := testGitHubClient()
		if _, err := c.CreateFile(IntegrationTestOwner, tc.repo, tc.branch, tc.path, tc.message, []byte(tc.content)); err == nil {
			t.Fatalf("#%d CreateFile: error is not supposed to be nil", i)
		}
	}
}

func TestGitHubClient_DeleteFileFail(t *testing.T) {
	c := testGitHubClient()
	f, err := c.GetFile(IntegrationTestOwner, IntegrationTestRepo, "master", "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	cases := []struct {
		repo, branch, path, sha, message string
	}{
		{repo: "unknown", branch: "master", path: "main.go", sha: *f.SHA, message: "test!"},
		{repo: TestRepo, branch: "unknown", path: "main.go", sha: *f.SHA, message: "test!"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: "unknown", message: "test!"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: "unknown", message: "test!"},
	}

	for i, tc := range cases {
		if err := c.DeleteFile(IntegrationTestOwner, tc.repo, tc.branch, tc.path, tc.sha, tc.message); err == nil {
			t.Fatalf("#%d DeleteFile: error is not supposed to be nil", i)
		}
	}
}

func TestGitHubClient_CreateAndDeleteFile(t *testing.T) {
	c := testGitHubClient()

	// Create a test branch
	branch := "test_create_and_delete_file"
	if err := c.CreateBranch(IntegrationTestOwner, IntegrationTestRepo, "master", branch); err != nil {
		t.Fatalf("unexpected error occured while creating a test branch: %s", err)
	}

	defer c.DeleteLatestRef(IntegrationTestOwner, IntegrationTestRepo, branch)

	// Create a file
	rc, err := c.CreateFile(IntegrationTestOwner, IntegrationTestRepo, branch, "new_file.go", "This is a new file!", []byte(`fmt.Println("test!")`))

	if err != nil {
		t.Fatalf("unexpected error occured while creating a file: %s", err)
	}

	// Be nice to GitHub API
	time.Sleep(3 * time.Second)

	// Delete a file
	if err := c.DeleteFile(IntegrationTestOwner, IntegrationTestRepo, branch, "new_file.go", *rc.Content.SHA, "This is a new file!"); err != nil {
		t.Fatalf("unexpected error occured while Deleting a file: %s", err)
	}
}

func TestUpdateFileFail(t *testing.T) {
	c := testGitHubClient()
	f, err := c.GetFile(IntegrationTestOwner, IntegrationTestRepo, "master", "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	cases := []struct {
		repo, branch, path, sha, message, content string
	}{
		{repo: "unknown", branch: "master", path: "main.go", sha: *f.SHA, message: "test!", content: "test"},
		{repo: TestRepo, branch: "unknown", path: "main.go", sha: *f.SHA, message: "test!", content: "test"},
		{repo: TestRepo, branch: "master", path: "main.go", sha: "unknown", message: "test!", content: "test"},
	}

	for i, tc := range cases {
		if err := c.UpdateFile(IntegrationTestOwner, tc.repo, tc.branch, tc.path, tc.sha, tc.message, []byte(tc.content)); err == nil {
			t.Fatalf("#%d UpdateFile: error is not supposed to be nil", i)
		}
	}
}

func TestUpdateFileSuccess(t *testing.T) {
	c := testGitHubClient()

	testBranch := "test_update_file"

	// Create new branches for this test
	err := c.CreateBranch(IntegrationTestOwner, IntegrationTestRepo, "master", testBranch)

	if err != nil {
		t.Fatalf("CreateBranch: unexpected error occured: %s", err)
	}

	// Delete the branches created for this test
	defer func() {
		err = c.DeleteLatestRef(IntegrationTestOwner, IntegrationTestRepo, testBranch)
		if err != nil {
			t.Fatalf("DeleteLatestRef: unexpected error occured: %s", err)
		}
	}()

	// Get main.go on the test branch
	f, err := c.GetFile(IntegrationTestOwner, IntegrationTestRepo, testBranch, "main.go")

	if err != nil {
		t.Fatalf("GetFile: unexpected error occured: %s", err)
	}

	// Update main.go on the test branch
	err = c.UpdateFile(IntegrationTestOwner, IntegrationTestRepo, testBranch, "main.go", *f.SHA, "Update main.go", []byte("test!"))

	if err != nil {
		t.Fatalf("UpdateFile: unexpected error occured: %s", err)
	}

	// Get updated main.go on the test branch
	f, err = c.GetFile(IntegrationTestOwner, IntegrationTestRepo, testBranch, "main.go")

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

func TestGitHubClient_DeleteRepositoryFail(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "unknown"},
	}

	for i, tc := range cases {
		c := testGitHubClient()

		if err := c.DeleteRepository(IntegrationTestOwner, tc.name); err == nil {
			t.Fatalf("#%d DeleteRepository: error is not supposed to be nil", i)
		}
	}
}

func TestGitHubClient_CreateAndDeleteRepository(t *testing.T) {
	c := testGitHubClient()
	repo := "test_create_repo"

	// Create Repo
	_, err := c.CreateRepository("", repo, "This is a test!", "", false)

	if err != nil {
		t.Fatalf("unexpected error occured while creating a GitHub repository: %s", err)
	}

	// Be nice to GitHub API
	time.Sleep(3 * time.Second)

	// Delete Repo
	err = c.DeleteRepository(IntegrationTestOwner, repo)

	if err != nil {
		t.Fatalf("unexpected error occured while deleting a GitHub repository: %s", err)
	}
}
