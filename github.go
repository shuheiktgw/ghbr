package go_homebrew

import (
	"golang.org/x/oauth2"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"context"
	"net/http"
)

// GitHubClient is a clint to interact with Github API
type GitHubClient struct {
	Owner string
	Client *github.Client
}

// NewGitHubClient creates and initializes a new GitHubClient
func NewGitHubClient(owner, token string) (*GitHubClient, error) {
	if len(owner) == 0 {
		return nil, errors.New("missing Github owner name")
	}

	if len(token) == 0 {
		return nil, errors.New("missing Github API token")
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
	rr, res, err := g.Client.Repositories.GetLatestRelease(context.TODO(), g.Owner, repo)

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("get latest version: invalid status: %s", res.Status)
	}

	return rr, err
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
		return nil, errors.Errorf("create pull request: invalid status: %s", res.Status)
	}

	return pr, nil
}

// MergePullRequest merges Pull Request with a give Pull Request number
func (g *GitHubClient) MergePullRequest(repo string, number int) error {
	_, res, err := g.Client.PullRequests.Merge(context.TODO(), g.Owner, repo, number, "", nil)

	if err != nil {
		return errors.Wrap(err, "failed to merge Pull Request")
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("merge Pull Request: invalid status: %s", res.Status)
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
		return errors.Errorf("close Pull Request: invalid status: %s", res.Status)
	}

	return nil
}

// UpdateFile updates a file with a given content
func (g *GitHubClient) UpdateFile(repo, path, message, sha, branch string, content []byte) error {
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
		return errors.Errorf("update file: invalid status: %s", res.Status)
	}

	return nil
}




