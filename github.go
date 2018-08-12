package main

import (
	"golang.org/x/oauth2"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"context"
	"net/http"
	)

// TODO: Move this const to a properer place
const EnvGitHubToken = "GITHUB_TOKEN"

// GitHub defines functions to interact with GitHub API
type GitHub interface {
	GetLatestRelease(repo string) (*github.RepositoryRelease, error)
	CreateBranch(repo, origin, new string) error
	DeleteLatestRef(repo, branch string) error
	CreatePullRequest(repo, title, head, base, body string) (*github.PullRequest, error)
	MergePullRequest(repo string, number int) error
	ClosePullRequest(repo string, number int) error
	GetFile(repo, branch, path string) (*github.RepositoryContent, error)
	UpdateFile(repo, branch, path, sha, message string, content []byte) error
}

// GitHubClient is a clint to interact with Github API
type GitHubClient struct {
	Owner string
	Client *github.Client
}

// NewGitHubClient creates and initializes a new GitHubClient
func NewGitHubClient(owner, token string) (GitHub, error) {
	if len(owner) == 0 {
		return nil, errors.New("missing Github owner name")
	}

	if len(token) == 0 {
		return nil, errors.New("missing Github personal access token")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(context.TODO(), ts)

	client := github.NewClient(tc)

	return &GitHubClient{
		Owner: owner,
		Client: client,
	}, nil
}

// GetLatestRelease returns the latest release of the given Repository
func (g *GitHubClient) GetLatestRelease(repo string) (*github.RepositoryRelease, error) {
	if len(repo) == 0 {
		return nil, errors.New("missing Github repository name")
	}

	rr, res, err := g.Client.Repositories.GetLatestRelease(context.TODO(), g.Owner, repo)

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("get latest version: invalid http status: %s", res.Status)
	}

	return rr, err
}

// CreateNewBranch creates a new branch from the heads of the origin
func (g *GitHubClient) CreateBranch(repo, origin, new string) error {
	if len(repo) == 0 {
		return errors.New("missing Github repository name")
	}

	if len(origin) == 0 {
		return errors.New("missing Github origin branch name")
	}

	if len(new) == 0 {
		return errors.New("missing Github new branch name")
	}

	originRef, res, err := g.Client.Git.GetRef(context.TODO(), g.Owner, repo, "heads/" + origin)

	if err != nil {
		return errors.Wrapf(err, "failed to GetRef: branch name: %s", origin)
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("failed to GetRef: branch name: %s, invalid http status: %s", res.Status)
	}

	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + new),
		Object: &github.GitObject{
			SHA: originRef.Object.SHA,
		},
	}

	_, res, err = g.Client.Git.CreateRef(context.TODO(), g.Owner, repo, newRef)

	if err != nil {
		return errors.Wrap(err, "failed to CreateRef")
	}

	if res.StatusCode != http.StatusCreated {
		return errors.Errorf("CreateRef: invalid http status: %s", res.Status)
	}

	return nil
}

// DeleteLatestRef deletes the latest Ref of the given branch, intended to be used for rollbacks
func (g *GitHubClient) DeleteLatestRef(repo, branch string) error {
	if len(repo) == 0 {
		return errors.New("missing Github repository name")
	}

	if len(branch) == 0 {
		return errors.New("missing Github branch name")
	}

	res, err := g.Client.Git.DeleteRef(context.TODO(), g.Owner, repo, "heads/" + branch)

	if err != nil {
		return errors.Wrapf(err, "failed to DeleteRef of a branch name %s: %s", branch, err)
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.Errorf("DeleteRef: invalid http status: %s", res.Status)
	}

	return nil
}

// CreatePullRequest creates Pull Request
func (g *GitHubClient) CreatePullRequest(repo, title, head, base, body string) (*github.PullRequest, error) {
	if len(title) == 0 {
		return nil, errors.New("missing Github Pull Request title")
	}

	if len(head) == 0 {
		return nil, errors.New("missing Github Pull Request head branch")
	}

	if len(base) == 0 {
		return nil, errors.New("missing Github Pull Request base branch")
	}

	if len(body) == 0 {
		return nil, errors.New("missing Github Pull Request body")
	}

	opt := &github.NewPullRequest{Title: &title, Head: &head, Base: &base, Body: &body}

	pr, res, err := g.Client.PullRequests.Create(context.TODO(), g.Owner, repo, opt)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new pull request")
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Errorf("create pull request: invalid http status: %s", res.Status)
	}

	return pr, nil
}

// MergePullRequest merges Pull Request with a give Pull Request number
func (g *GitHubClient) MergePullRequest(repo string, number int) error {
	if len(repo) == 0 {
		return errors.New("missing Github Repository name")
	}

	if number == 0 {
		return errors.New("missing Github Pull Request number")
	}

	_, res, err := g.Client.PullRequests.Merge(context.TODO(), g.Owner, repo, number, "", nil)

	if err != nil {
		return errors.Wrap(err, "failed to merge Pull Request")
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("merge Pull Request: invalid http status: %s", res.Status)
	}

	return nil
}

// ClosePullRequest closes Pull Request with a give Pull Request number
func (g *GitHubClient) ClosePullRequest(repo string, number int) error {
	opt := &github.PullRequest{State: github.String("close")}

	_, res, err := g.Client.PullRequests.Edit(context.TODO(), g.Owner, repo, number, opt)

	if err != nil {
		return errors.Wrap(err, "failed to close Pull Request")
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("close Pull Request: invalid http status: %s", res.Status)
	}

	return nil
}

// GetFile gets the specified file on GitHub
func (g *GitHubClient) GetFile(repo, branch, path string) (*github.RepositoryContent, error) {
	if len(repo) == 0 {
		return nil, errors.New("missing Github repository name")
	}
	
	if len(branch) == 0 {
		return nil, errors.New("missing Github branch name")
	}

	if len(path) == 0 {
		return nil, errors.New("missing Github file path")
	}

	opt := &github.RepositoryContentGetOptions{Ref: branch}

	file, _, res, err := g.Client.Repositories.GetContents(context.TODO(), g.Owner, repo, path, opt)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to GetVersion: repository name: %s, branch name: %s, file path: %s", repo, branch, path)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("GetFile: invalid http status: %s", res.Status)
	}

	return file, nil
}

// UpdateFile updates a file with a given content
func (g *GitHubClient) UpdateFile(repo, branch, path, sha, message string, content []byte) error {
	if len(path) == 0 {
		return errors.New("missing Github file path")
	}

	if len(message) == 0 {
		return errors.New("missing Github commit message")
	}

	if len(content) == 0 {
		return errors.New("missing Github content")
	}

	if len(sha) == 0 {
		return errors.New("missing Github file sha")
	}

	if len(branch) == 0 {
		return errors.New("missing Github branch name")
	}

	opt := &github.RepositoryContentFileOptions{Message: &message, Content: content, SHA: &sha, Branch: &branch}

	_, res, err := g.Client.Repositories.UpdateFile(context.TODO(), g.Owner, repo, path, opt)

	if err != nil {
		return errors.Wrap(err, "failed to update file")
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("update file: invalid http status: %s", res.Status)
	}

	return nil
}
