package main

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
	"github.com/google/go-github/github"
	"github.com/iancoleman/strcase"
	"github.com/pkg/errors"
)

const (
	CreateBranch = "master"
)

var versionRegex = regexp.MustCompile(`version\s['"]([\w.-]+)['"]`)
var urlRegex = regexp.MustCompile(`url\s['"]((http|https)://[\w-./?%&=]+)['"]`)
var shaRegex = regexp.MustCompile(`sha256\s['"]([0-9A-Fa-f]{64})['"]`)

// HandledError represents the error is properly handled and is not unexpected
type HandledError struct {
	Message string
}

func (e *HandledError) Error() string {
	return e.Message
}

// Ghbr defines functions for Homebrew Formula
type Ghbr struct {
	GitHub *GitHubClient

	outStream io.Writer
}

// LatestRelease contains latest release info
type LatestRelease struct {
	version, url, hash string
}

// GetLatestRelease returns the latest release and calculates its checksum
func (g *Ghbr) GetLatestRelease(owner, repo string) (*LatestRelease, error) {
	// Get latest release of the repository
	fmt.Fprint(g.outStream, "[ghbr] ===> Checking the latest release\n")
	release, err := g.GitHub.GetLatestRelease(owner, repo)
	if err != nil {
		return nil, err
	}

	// Extract version
	version := *release.TagName

	// Get a URL of a released asset for Mac
	url, err := findMacAssetURL(release)
	if err != nil {
		return nil, err
	}

	// Download the release asset
	fmt.Fprint(g.outStream, "[ghbr] ===> Downloading Darwin AMD64 release\n")
	body, err := g.downloadFile(url)
	if err != nil {
		return nil, err
	}

	defer body.Close()

	// Calculate hash
	fmt.Fprint(g.outStream, "[ghbr] ===> Calculating a checksum of the release\n")
	hash, err := calculateSha256(body)
	if err != nil {
		return nil, err
	}

	return &LatestRelease{version: version, url: url, hash: hash}, nil
}

func (g *Ghbr) CreateFormula(org, owner, app, font string, private bool, release *LatestRelease) error {
	// Create a new Repository
	formulaRepoName := fmt.Sprintf("homebrew-%s", app)
	originalRepo := fmt.Sprintf("%s/%s", owner, app)

	fmt.Fprint(g.outStream, "[ghbr] ===> Creating a repository\n")
	repo, err := g.GitHub.CreateRepository(
		org,
		formulaRepoName,
		fmt.Sprintf("Homebrew formula for %s", originalRepo),
		fmt.Sprintf("https://github.com/%s", originalRepo),
		private,
	)

	if err != nil {
		return err
	}

	var formulaOwner string
	if len(org) != 0 {
		formulaOwner = org
	} else {
		formulaOwner = owner
	}

	// Create README.md
	fmt.Fprint(g.outStream, "[ghbr] ===> Adding README.md to the repository\n")
	if err := g.createReadme(formulaOwner, formulaRepoName, originalRepo); err != nil {
		return err
	}

	// Create Formula
	fmt.Fprintf(g.outStream, "[ghbr] ===> Adding %s.rb to the repository\n", app)
	if err := g.createFormula(formulaOwner, app, formulaRepoName, originalRepo, font, release); err != nil {
		return err
	}

	fmt.Fprintf(g.outStream, "\n\n")
	fmt.Fprintf(g.outStream, "Yay! Your Homebrew formula repository has been successfully created!\n")
	fmt.Fprintf(g.outStream, "Access %s and see what we achieved.\n\n", *repo.HTMLURL)

	return nil
}

// UpdateFormula updates the formula file to point to the latest release
func (g *Ghbr) UpdateFormula(org, owner, app, branch string, force, merge bool, release *LatestRelease) error {
	repo := fmt.Sprintf("homebrew-%s", app)
	path := fmt.Sprintf("%s.rb", app)

	var formulaOwner string
	if len(org) != 0 {
		formulaOwner = org
	} else {
		formulaOwner = owner
	}

	// Get the formula file
	fmt.Fprintf(g.outStream, "[ghbr] ===> Checking the current formula\n")
	rc, err := g.GitHub.GetFile(formulaOwner, repo, branch, path)

	if err != nil {
		return err
	}

	// Decode the formula file
	currentFormula, err := decodeContent(rc)

	if err != nil {
		return err
	}

	// Check current version
	upToDate, err := checkVersionLatest(currentFormula, release)

	if err != nil {
		return err
	}

	if upToDate && !force {
		fmt.Fprintf(g.outStream, "\n\n")
		fmt.Fprintf(g.outStream, "ghbr aborted!\n\n")

		fmt.Fprintf(g.outStream, "The current formula (pointing to version %s) is up-to-date.\n", release.version)
		fmt.Fprintf(g.outStream, "If you want to update the formula anyway, run `ghbr release` with `--force` option.\n\n")
		return nil
	}

	// Edit the formula file
	newFormula, err := bumpsUpFormula(currentFormula, release)

	if err != nil {
		return err
	}

	// Create a new feature branch
	fmt.Fprintf(g.outStream, "[ghbr] ===> Creating a new feature branch\n")
	newBranch := fmt.Sprintf("bumps_up_to_%s", release.version)

	err = g.GitHub.CreateBranch(formulaOwner, repo, branch, newBranch)

	if err != nil {
		return err
	}

	// Update formula file on the feature branch
	fmt.Fprintf(g.outStream, "[ghbr] ===> Updating the formula file\n")
	message := fmt.Sprintf("Bumps up to %s", release.version)

	err = g.GitHub.UpdateFile(
		formulaOwner,
		repo,
		newBranch,
		path,
		*rc.SHA,
		message,
		[]byte(newFormula),
	)

	if err != nil {
		// Delete branch if the update fails
		g.GitHub.DeleteLatestRef(formulaOwner, repo, newBranch)

		return err
	}

	// Create a PR from the feature branch to its origin
	fmt.Fprintf(g.outStream, "[ghbr] ===> Creating a Pull Request\n")
	pr, err := g.GitHub.CreatePullRequest(formulaOwner, repo, message, newBranch, branch, message)

	if err != nil {
		// Delete the branch if PR creation fails
		g.GitHub.DeleteLatestRef(formulaOwner, repo, newBranch)

		return err
	}

	// Merge the PR
	if merge {
		fmt.Fprintf(g.outStream, "[ghbr] ===> Merging the Pull Request\n")

		if err := g.GitHub.MergePullRequest(formulaOwner, repo, *pr.Number); err != nil {
			// Delete the branch and the PR if the merge fails]
			g.GitHub.ClosePullRequest(formulaOwner, repo, *pr.Number)
			g.GitHub.DeleteLatestRef(formulaOwner, repo, newBranch)

			return err
		}

		fmt.Fprintf(g.outStream, "[ghbr] ===> Deleting the branch\n")

		if err := g.GitHub.DeleteLatestRef(formulaOwner, repo, newBranch); err != nil {
			return err
		}

		fmt.Fprintf(g.outStream, "\n\n")
		fmt.Fprintf(g.outStream, "Yay! Now your formula is up-to-date!\n\n")

		return nil
	}

	fmt.Fprintf(g.outStream, "\n\n")
	fmt.Fprintf(g.outStream, "Yay! Now your formula is ready to update!\n\n")
	fmt.Fprintf(g.outStream, "Access %s and merge the Pull Request\n\n", *pr.HTMLURL)

	return nil
}

// createReadme creates a README.md on master branch
func (g *Ghbr) createReadme(owner, formulaRepoName, originalRepo string) error {

	content := fmt.Sprintf(`%s
====

[Homebrew](http://brew.sh/) formula for [%s](https://github.com/%s)

`, formulaRepoName, originalRepo, originalRepo)

	_, err := g.GitHub.CreateFile(
		owner,
		formulaRepoName,
		"master",
		"README.md",
		"Create README.md",
		[]byte(content),
	)

	return err
}

// createReadme creates a README.md on master branch
func (g *Ghbr) createFormula(owner, app, repo, originalRepo, font string, release *LatestRelease) error {

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
end

`, strcase.ToCamel(app), originalRepo, release.version, release.url, release.hash, app, caveats.String())

	_, err := g.GitHub.CreateFile(
		owner,
		repo,
		CreateBranch,
		fmt.Sprintf("%s.rb", app),
		"Create formula",
		[]byte(content),
	)

	return err
}

// downloadFile downloads a file from the url and return the content
func (g *Ghbr) downloadFile(url string) (io.ReadCloser, error) {
	// Get the data
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("#downloadFile returns invalid http status: %s", res.Status)
	}

	return res.Body, nil
}

func findMacAssetURL(release *github.RepositoryRelease) (string, error) {
	for _, a := range release.Assets {
		if strings.Contains(*a.Name, "darwin") && strings.Contains(*a.Name, "amd64") {
			return *a.BrowserDownloadURL, nil
		}
	}

	return "", &HandledError{
		Message: `No released asset whose name contains "darwin" and "amd64".\n` +
			`You need to name one of the assets with "darwin" and "amd64" to specify the asset is for Mac.`,
	}
}

func calculateSha256(content io.ReadCloser) (string, error) {
	sha := sha256.New()

	if _, err := io.Copy(sha, content); err != nil {
		return "", err
	}

	return hex.EncodeToString(sha.Sum(nil)), nil
}

func decodeContent(rc *github.RepositoryContent) (string, error) {
	if *rc.Encoding != "base64" {
		return "", errors.Errorf("unexpected encoding: %s", *rc.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(*rc.Content)

	if err != nil {
		return "", errors.Wrap(err, "error occurred while decoding version.rb file")
	}

	return string(decoded), nil
}

func checkVersionLatest(content string, release *LatestRelease) (bool, error) {
	ms := versionRegex.FindStringSubmatch(content)

	if ms == nil {
		return false, errors.New("could not find version in a formula file")
	}

	return ms[1] == release.version, nil
}

func bumpsUpFormula(content string, release *LatestRelease) (string, error) {
	// Update version
	c, err := findAndReplace(versionRegex, content, release.version)

	if err != nil {
		return "", errors.Wrap(err, "formula file is likely not to contain proper `version` indicator")
	}

	// Update url
	c, err = findAndReplace(urlRegex, c, release.url)

	if err != nil {
		return "", errors.Wrap(err, "formula file is likely not to contain proper `url` indicator")
	}

	// Update hash
	c, err = findAndReplace(shaRegex, c, release.hash)
	if err != nil {
		return "", errors.Wrap(err, "formula file is likely not to contain proper `sha256` indicator")
	}

	return c, nil
}

func findAndReplace(reg *regexp.Regexp, content, new string) (string, error) {
	ms := reg.FindStringSubmatch(content)

	if ms == nil {
		return "", &HandledError{Message: fmt.Sprintf("could not find sub match: content: %s, regex: %v", content, reg)}
	}

	old := ms[1]

	return strings.Replace(content, old, new, -1), nil
}
