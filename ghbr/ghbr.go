package ghbr

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/shuheiktgw/ghbr/github"
	goGithub "github.com/google/go-github/github"
	"github.com/pkg/errors"
)

const MacRelease = "darwin_amd64"

var versionRegex = regexp.MustCompile(`version\s['"]([\w.-]+)['"]`)
var urlRegex = regexp.MustCompile(`url\s['"]((http|https)://[\w-./?%&=]+)['"]`)
var shaRegex = regexp.MustCompile(`sha256\s['"]([0-9A-Fa-f]{64})['"]`)

// GHBR defines GHBRClient interface
type GHBR interface {
	GetCurrentRelease(repo string) (*LatestRelease, error)
	UpdateFormula(app, branch string, release *LatestRelease)
	DownloadFile(path, url string) error
}

// GHBRClient define functions for Homebrew Formula
type GHBRClient struct {
	GitHub github.GitHub

	outStream io.Writer
}

// LatestRelease contains latest release info
type LatestRelease struct {
	version, url, hash string
}

// GetCurrentRelease returns the latest release version and calculates its checksum
func (g *GHBRClient) GetCurrentRelease(repo string) (*LatestRelease, error) {
	if len(repo) == 0 {
		return nil, errors.New("missing GitHub repository")
	}

	// Get latest release of the repository
	release, err := g.GitHub.GetLatestRelease(repo)

	if err != nil {
		return nil, err
	}

	// Extract version
	version := *release.TagName

	// Get Mac release asset browser URL
	url, err := findMacReleaseURL(release)

	// Download the release asset
	path := "darwin_amd64.zip"
	err = g.DownloadFile(path, url)
	defer os.Remove(path)

	if err != nil {
		return nil, err
	}

	// Calculate hash
	hash, err := calculateSha256(path)

	if err != nil {
		return nil, err
	}

	return &LatestRelease{version: version, url: url, hash: hash}, nil
}

// UpdateFormula updates the formula file to point to the latest release
func (g *GHBRClient) UpdateFormula(app, branch string, release *LatestRelease) error {
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
	newBranch := fmt.Sprintf("bumps_up_to_%s", release.version)

	err = g.GitHub.CreateBranch(repo, branch, newBranch)

	if err != nil {
		return err
	}

	// Update formula file on the feature branch
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
	_, err = g.GitHub.CreatePullRequest(repo, message, newBranch, branch, message)

	if err != nil {
		// Delete branch if the create branch fails
		g.GitHub.DeleteLatestRef(repo, newBranch)

		return err
	}

	return nil

}

// DownloadFile downloads a file from the url and save it to the path
func (g *GHBRClient) DownloadFile(path, url string) error {
	if len(path) == 0 {
		return errors.New("missing download file path")
	}

	if len(url) == 0 {
		return errors.New("missing download url")
	}

	// Create the file
	out, err := os.Create(path)

	if err != nil {
		return err
	}

	defer out.Close()

	// Get the data
	res, err := http.Get(url)

	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.Errorf("DownloadFile: invalid http status: %s", res.Status)
	}

	defer res.Body.Close()

	// Write the body to the file
	if _, err = io.Copy(out, res.Body); err != nil {
		return err
	}

	return nil
}

func findMacReleaseURL(release *goGithub.RepositoryRelease) (string, error) {
	for _, a := range release.Assets {
		if strings.Contains(*a.Name, MacRelease) {
			return *a.BrowserDownloadURL, nil
		}
	}

	return "", errors.Errorf("could not find assets named %s", MacRelease)
}

func calculateSha256(path string) (string, error) {
	sha := sha256.New()

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if _, err := io.Copy(sha, f); err != nil {
		return "", err
	}

	return string(sha.Sum(nil)), nil
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
