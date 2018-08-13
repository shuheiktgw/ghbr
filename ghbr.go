package main

import (
	"io"
	"os"
	"net/http"
	"github.com/pkg/errors"
	"github.com/google/go-github/github"
	"strings"
	"crypto/sha256"
	"fmt"
	"encoding/base64"
	"regexp"
)

const MacRelease = "darwin_amd64"

var versionRegex = regexp.MustCompile(`version\s['"]([\w.-]+)['"]`)
var urlRegex = regexp.MustCompile(`url\s['"]((http|https)://[\w-./?%&=]+)['"]`)
var shaRegex = regexp.MustCompile(`sha256\s['"]([0-9A-Fa-f]{64})['"]`)

// GHBR define functions for Homebrew Formula
type GHBR struct {
	GitHub GitHub

	outStream io.Writer
}

// LatestRelease contains latest release info
type LatestRelease struct {
	version, url, hash string
}

// GetLatestRelease returns the latest release version and calculates its checksum
func (g *GHBR) GetLatestRelease(repo string) (*LatestRelease, error) {
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
func (g *GHBR) UpdateFormula(app, branch string, release *LatestRelease) error {
	if len(app) == 0 {
		return errors.New("missing application name")
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
	_, err = decodeContent(rc)

	if err != nil {
		return err
	}

	// Update the formula file

	return nil

}

// DownloadFile downloads a file from the url and save it to the path
func (g *GHBR) DownloadFile(path, url string) error {
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

func findMacReleaseURL(release *github.RepositoryRelease) (string, error) {
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

	if _, err :=  io.Copy(sha, f); err != nil {
		return "", err
	}

	return string(sha.Sum(nil)), nil
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

func updateFormula(content string, release *LatestRelease) (string, error) {
	// Update version
	vs := versionRegex.FindStringSubmatch(content)

	if vs == nil {
		return "", errors.New("could not find version definition in formula file")
	}

	v := vs[1]

	c := strings.Replace(content, v, release.version, -1)

	// Update url
	us := urlRegex.FindStringSubmatch(content)

	if us == nil {
		return "", errors.New("could not find url definition in formula file")
	}

	u := us[1]

	strings.Replace(c, u, release.url, -1)

	// Update hash
	ss := shaRegex.FindStringSubmatch(content)

	if ss == nil {
		return "", errors.New("could not find hash definition in formula file")
	}

	s := ss[1]

	return strings.Replace(c, s, release.hash, -1), nil
}








