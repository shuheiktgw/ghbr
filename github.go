package main

import (
	"context"
	"time"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// GitHubClient is a clint to interact with Github API
type GitHubClient struct {
	Client *github.Client
}

// NewGitHubClient creates and initializes a new GitHubClient
func NewGitHubClient(token string) *GitHubClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(context.TODO(), ts)

	client := github.NewClient(tc)

	return &GitHubClient{Client: client}
}

// GetLatestRelease returns the latest release of the given Repository
func (g *GitHubClient) GetLatestRelease(owner, repo string) (*github.RepositoryRelease, error) {
	rr, _, err := g.Client.Repositories.GetLatestRelease(context.TODO(), owner, repo)

	if err != nil {
		return nil, errors.Wrapf(err, "#Repositories.GetLatestRelease failed: owner: %s, repo: %s", owner, repo)
	}

	return rr, err
}

// CreateBranch creates a new branch from the heads of the origin
func (g *GitHubClient) CreateBranch(owner, repo, origin, new string) error {
	originRef, _, err := g.Client.Git.GetRef(context.TODO(), owner, repo, "heads/"+origin)

	if err != nil {
		return errors.Wrapf(err, "#Git.GetRef failed: owner: %s, repo: %s", owner, repo)
	}

	newRef := &github.Reference{
		Ref: github.String("refs/heads/" + new),
		Object: &github.GitObject{
			SHA: originRef.Object.SHA,
		},
	}

	_, _, err = g.Client.Git.CreateRef(context.TODO(), owner, repo, newRef)

	if err != nil {
		return errors.Wrapf(err, "#Git.CreateRef failed: owner: %s, repo: %s, ref: %v", owner, repo, newRef)
	}

	return nil
}

// DeleteLatestRef deletes the latest Ref of the given branch, intended to be used for rollbacks
func (g *GitHubClient) DeleteLatestRef(owner, repo, branch string) error {
	_, err := g.Client.Git.DeleteRef(context.TODO(), owner, repo, "heads/"+branch)

	if err != nil {
		return errors.Wrapf(err, "#Git.DeleteRef failed to delete the latest ref of %s branch", branch)
	}

	return nil
}

// CreatePullRequest creates Pull Request
func (g *GitHubClient) CreatePullRequest(owner, repo, title, head, base, body string) (*github.PullRequest, error) {
	opt := &github.NewPullRequest{Title: &title, Head: &head, Base: &base, Body: &body}

	pr, _, err := g.Client.PullRequests.Create(context.TODO(), owner, repo, opt)

	if err != nil {
		return nil, err
	}

	return pr, nil
}

// MergePullRequest merges Pull Request with a give Pull Request number
func (g *GitHubClient) MergePullRequest(owner, repo string, number int) error {
	// Wait a few seconds to prevent GitHub API returns `405 Base branch was modified`
	// TODO Move this code to the client side
	time.Sleep(3 * time.Second)

	_, _, err := g.Client.PullRequests.Merge(context.TODO(), owner, repo, number, "", nil)

	if err != nil {
		return err
	}

	return nil
}

// ClosePullRequest closes Pull Request with a give Pull Request number
func (g *GitHubClient) ClosePullRequest(owner, repo string, number int) error {
	pr := &github.PullRequest{State: github.String("close")}

	_, _, err := g.Client.PullRequests.Edit(context.TODO(), owner, repo, number, pr)

	if err != nil {
		return errors.Wrapf(err, "#PullRequests.Edit failed to close Pull Request: owner: %s, repo: %s, number: %d", owner, repo, number)
	}

	return nil
}

// GetFile gets the specified file on GitHub
func (g *GitHubClient) GetFile(owner, repo, branch, path string) (*github.RepositoryContent, error) {
	opt := &github.RepositoryContentGetOptions{Ref: branch}

	file, _, _, err := g.Client.Repositories.GetContents(context.TODO(), owner, repo, path, opt)

	if err != nil {
		return nil, errors.Wrapf(err, "#Repositories.GetContents failed: repository name: %s, branch name: %s, file path: %s", repo, branch, path)
	}

	return file, nil
}

// CreateFile create a file with a given content on GitHub
func (g *GitHubClient) CreateFile(owner, repo, branch, path, message string, content []byte) (*github.RepositoryContentResponse, error) {
	opt := &github.RepositoryContentFileOptions{Message: &message, Content: content, Branch: &branch}

	rc, _, err := g.Client.Repositories.CreateFile(context.TODO(), owner, repo, path, opt)

	if err != nil {
		return nil, errors.Wrapf(err, "#Repositories.CreateFile failed: owner: %s, repo: %s, branch: %s, path: %s", owner, repo, branch, path)
	}

	return rc, nil
}

// UpdateFile updates a file on GitHub with a given content
func (g *GitHubClient) UpdateFile(owner, repo, branch, path, sha, message string, content []byte) error {
	opt := &github.RepositoryContentFileOptions{Message: &message, Content: content, SHA: &sha, Branch: &branch}

	_, _, err := g.Client.Repositories.UpdateFile(context.TODO(), owner, repo, path, opt)

	if err != nil {
		return errors.Wrapf(err, "#Repositories.UpdateFile failed: repo: %s, branch: %s, path: %s", repo, branch, path)
	}

	return nil
}

// DeleteFile deletes a file on GitHub
func (g *GitHubClient) DeleteFile(owner, repo, branch, path, sha, message string) error {
	opt := &github.RepositoryContentFileOptions{Message: &message, SHA: &sha, Branch: &branch}

	_, _, err := g.Client.Repositories.DeleteFile(context.TODO(), owner, repo, path, opt)

	if err != nil {
		return errors.Wrapf(err, "#Repositories.DeleteFile failed: repo: %s, branch: %s, path: %s", repo, branch, path)
	}

	return nil
}

// CreateRepository creates a new GitHub repository
func (g *GitHubClient) CreateRepository(org, name, description, homepage string, private bool) (*github.Repository, error) {
	opt := &github.Repository{
		Name:        &name,
		Description: &description,
		Homepage:    &homepage,
		Private:     &private,
	}

	repo, _, err := g.Client.Repositories.Create(context.TODO(), org, opt)

	if err != nil {
		return nil, errors.Wrapf(err, "#Repositories.Create failed: repository org: %s, name: %s", org, name)
	}

	return repo, nil
}

// DeleteRepository deletes a GitHub repository
func (g *GitHubClient) DeleteRepository(owner, name string) error {
	_, err := g.Client.Repositories.Delete(context.TODO(), owner, name)

	if err != nil {
		return errors.Wrapf(err, "#Repositories.Delete failed: repository name: %s", name)
	}

	return nil
}
