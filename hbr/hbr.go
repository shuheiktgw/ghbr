package hbr

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/common-nighthawk/go-figure"
	goGithub "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/shuheiktgw/ghbr/github"
)

const (
	CreateBranch = "master"
	MacRelease   = "darwin_amd64"
)

var versionRegex = regexp.MustCompile(`version\s['"]([\w.-]+)['"]`)
var urlRegex = regexp.MustCompile(`url\s['"]((http|https)://[\w-./?%&=]+)['"]`)
var shaRegex = regexp.MustCompile(`sha256\s['"]([0-9A-Fa-f]{64})['"]`)

// Generator is a type for method to generate GBHRWrapper
type Generator func(token, owner string) GHBRWrapper

// Generator defines a method to create GHBRWrapper
func GenerateGHBR(token, owner string) GHBRWrapper {
	gitHub, err := github.NewGitHubClient(owner, token)

	if err != nil {
		return &Wrapper{client: nil, err: err}
	}

	return &Wrapper{client: &Client{GitHub: gitHub}, err: nil}
}

// GHBRWrapper abstracts Wrapper's interface
type GHBRWrapper interface {
	GetCurrentRelease(repo string) *LatestRelease
	CreateFormula(app, font string, private bool, release *LatestRelease)
	UpdateFormula(app, branch string, merge bool, release *LatestRelease)
	Err() error
}

// Wrapper wraps Client to avoid ugly error handling
type Wrapper struct {
	client GHBRClient
	err    error
}

// GetCurrentRelease wraps client's GetCurrentRelease method
func (w *Wrapper) GetCurrentRelease(repo string) *LatestRelease {
	if w.err != nil {
		return nil
	}

	lr, err := w.client.GetCurrentRelease(repo)

	w.err = err

	return lr
}

// CreateFormula wraps client's CreateFormula method
func (w *Wrapper) CreateFormula(app, font string, private bool, release *LatestRelease) {
	if w.err != nil {
		return
	}

	w.err = w.client.CreateFormula(app, font, private, release)

	return
}

// UpdateFormula wraps client's UpdateFormula method
func (w *Wrapper) UpdateFormula(app, branch string, merge bool, release *LatestRelease) {
	if w.err != nil {
		return
	}

	w.err = w.client.UpdateFormula(app, branch, merge, release)

	return
}

// Err returns the value of the err filed
func (w *Wrapper) Err() error {
	return w.err
}

// GHBRClient abstracts Client's interface
type GHBRClient interface {
	GetCurrentRelease(repo string) (*LatestRelease, error)
	CreateFormula(app, font string, private bool, release *LatestRelease) error
	UpdateFormula(app, branch string, merge bool, release *LatestRelease) error
}

// Client define functions for Homebrew Formula
type Client struct {
	GitHub github.GitHub
}

// LatestRelease contains latest release info
type LatestRelease struct {
	version, url, hash string
}

// GetCurrentRelease returns the latest release version and calculates its checksum
func (g *Client) GetCurrentRelease(repo string) (*LatestRelease, error) {
	if len(repo) == 0 {
		return nil, errors.New("missing GitHub repository")
	}

	// Get latest release of the repository
	fmt.Println("===> Getting the latest release")
	release, err := g.GitHub.GetLatestRelease(repo)

	if err != nil {
		return nil, err
	}

	// Extract version
	version := *release.TagName

	// Get Mac release asset browser URL
	url, err := findMacReleaseURL(release)

	// Download the release asset
	fmt.Println("===> Downloading Darwin AMD64 release")
	body, err := g.downloadFile(url)

	if err != nil {
		return nil, err
	}

	defer body.Close()

	// Calculate hash
	fmt.Println("===> Calculating a checksum of the release")
	hash, err := calculateSha256(body)

	if err != nil {
		return nil, err
	}

	return &LatestRelease{version: version, url: url, hash: hash}, nil
}

func (g *Client) CreateFormula(app, font string, private bool, release *LatestRelease) error {
	if len(app) == 0 {
		return errors.New("missing application name")
	}

	if len(font) == 0 {
		return errors.New("missing ascii font name")
	}

	if release == nil {
		return errors.New("missing GitHub release")
	}

	// Create a new Repository
	formulaRepoName := fmt.Sprintf("homebrew-%s", app)
	originalRepo := fmt.Sprintf("%s/%s", g.GitHub.GetOwner(), app)

	err := g.GitHub.CreateRepository(
		formulaRepoName,
		fmt.Sprintf("Homebrew formula for %s", originalRepo),
		fmt.Sprintf("https://github.com/%s", originalRepo),
		private,
	)

	if err != nil {
		return err
	}

	// Create README.md
	if err := g.createReadme(formulaRepoName, originalRepo); err != nil {
		return err
	}

	// Create Formula
	if err := g.createFormula(app, formulaRepoName, originalRepo, font, release); err != nil {
		return err
	}

	return nil
}

// UpdateFormula updates the formula file to point to the latest release
func (g *Client) UpdateFormula(app, branch string, merge bool, release *LatestRelease) error {
	if len(app) == 0 {
		return errors.New("missing application name")
	}

	if len(branch) == 0 {
		return errors.New("missing GitHub branch")
	}

	if release == nil {
		return errors.New("missing GitHub release")
	}

	repo := fmt.Sprintf("homebrew-%s", app)
	path := fmt.Sprintf("%s.rb", app)

	// Get the formula file
	fmt.Println("===> Getting the current formula file")
	rc, err := g.GitHub.GetFile(repo, branch, path)

	if err != nil {
		return err
	}

	// Decode the formula file
	oldFormula, err := decodeContent(rc)

	if err != nil {
		return err
	}

	// Edit the formula file
	newFormula, err := bumpsUpFormula(oldFormula, release)

	if err != nil {
		return err
	}

	// Create a new feature branch
	fmt.Println("===> Creating a new branch")
	newBranch := fmt.Sprintf("bumps_up_to_%s", release.version)

	err = g.GitHub.CreateBranch(repo, branch, newBranch)

	if err != nil {
		return err
	}

	// Update formula file on the feature branch
	fmt.Println("===> Updating the formula file")
	message := fmt.Sprintf("Bumps up to %s", release.version)

	err = g.GitHub.UpdateFile(
		repo,
		newBranch,
		path,
		*rc.SHA,
		message,
		[]byte(newFormula),
	)

	if err != nil {
		// Delete branch if the update fails
		g.GitHub.DeleteLatestRef(repo, newBranch)

		return err
	}

	// Create a PR from the feature branch to its origin
	fmt.Println("===> Creating a Pull Request")
	pr, err := g.GitHub.CreatePullRequest(repo, message, newBranch, branch, message)

	if err != nil {
		// Delete the branch if PR creation fails
		g.GitHub.DeleteLatestRef(repo, newBranch)

		return err
	}

	// Merge the PR
	if merge {
		fmt.Println("===> Merging the Pull Request")

		if err := g.GitHub.MergePullRequest(repo, *pr.Number); err != nil {
			// Delete the branch and the PR if the merge fails]
			g.GitHub.ClosePullRequest(repo, *pr.Number)
			g.GitHub.DeleteLatestRef(repo, newBranch)

			return err
		}

		fmt.Printf("\n\n")
		fmt.Printf("Yay! Now your formula is up-to-date!\n\n")

		return nil
	}

	fmt.Printf("\n\n")
	fmt.Printf("Yay! Now your formula is ready to update!\n\n")
	fmt.Printf("Remaining tasks are below:\n")
	fmt.Printf("Access %s and merge the Pull Request\n\n", *pr.HTMLURL)

	return nil
}

// createReadme creates a README.md on master branch
func (g *Client) createReadme(formulaRepoName, originalRepo string) error {

	content := fmt.Sprintf(`%s
====

[Homebrew](http://brew.sh/) formula for [%s](https://github.com/%s)

`, formulaRepoName, originalRepo, originalRepo)

	_, err := g.GitHub.CreateFile(
		formulaRepoName,
		CreateBranch,
		"README.md",
		"Create README.md",
		[]byte(content),
	)

	return err
}

// createReadme creates a README.md on master branch
func (g *Client) createFormula(app, repo, originalRepo, font string, release *LatestRelease) error {

	caveats := figure.NewFigure(app, font, true)

	content := fmt.Sprintf(`require 'formula'

class %s < Formula
  homepage 'https://github.com/%s'
  version '%s'

  url '%s'
  sha256 '%s'

  def install
    bin.install '%s'
  end

  def caveats
    <<-'EOF'
%s
    EOF
end

`, strings.Title(app), originalRepo, release.version, release.url, release.hash, app, caveats.String())

	_, err := g.GitHub.CreateFile(
		repo,
		CreateBranch,
		fmt.Sprintf("%s.rb", app),
		"Create formula",
		[]byte(content),
	)

	return err
}

// downloadFile downloads a file from the url and return the content
func (g *Client) downloadFile(url string) (io.ReadCloser, error) {
	if len(url) == 0 {
		return nil, errors.New("missing download url")
	}

	// Get the data
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("downloadFile: invalid http status: %s", res.Status)
	}

	return res.Body, nil
}

func findMacReleaseURL(release *goGithub.RepositoryRelease) (string, error) {
	for _, a := range release.Assets {
		if strings.Contains(*a.Name, MacRelease) {
			return *a.BrowserDownloadURL, nil
		}
	}

	return "", errors.Errorf("could not find assets named %s", MacRelease)
}

func calculateSha256(content io.ReadCloser) (string, error) {
	sha := sha256.New()

	if _, err := io.Copy(sha, content); err != nil {
		return "", err
	}

	return hex.EncodeToString(sha.Sum(nil)), nil
}

func decodeContent(rc *goGithub.RepositoryContent) (string, error) {
	if *rc.Encoding != "base64" {
		return "", errors.Errorf("unexpected encoding: %s", *rc.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(*rc.Content)

	if err != nil {
		return "", errors.Wrap(err, "error occurred while decoding version.rb file")
	}

	return string(decoded), nil
}

func bumpsUpFormula(content string, release *LatestRelease) (string, error) {
	// Update version
	c, err := findAndReplace(versionRegex, content, release.version)

	if err != nil {
		return "", err
	}

	// Update url
	c, err = findAndReplace(urlRegex, c, release.url)

	if err != nil {
		return "", err
	}

	// Update hash
	return findAndReplace(shaRegex, c, release.hash)
}

func findAndReplace(reg *regexp.Regexp, content, new string) (string, error) {
	ms := reg.FindStringSubmatch(content)

	if ms == nil {
		return "", errors.New("could not find submatch")
	}

	old := ms[1]

	return strings.Replace(content, old, new, -1), nil
}
