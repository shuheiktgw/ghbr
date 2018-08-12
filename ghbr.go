package main

import (
	"io"
	"os"
	"net/http"
	"github.com/pkg/errors"
	"github.com/google/go-github/github"
	"strings"
	"crypto/sha256"
)

const MacRelease = "darwin_amd64"

// GHBR define functions for Homebrew Formula
type GHBR struct {
	GitHub GitHub

	outStream io.Writer
}

// LatestRelease contains latest release info
type LatestRelease struct {
	tag, hash string
}

// GetLatestRelease returns the latest release tag and calculates its checksum
func (g *GHBR) GetLatestRelease(repo string) (*LatestRelease, error) {
	if len(repo) == 0 {
		return nil, errors.New("missing GitHub repository")
	}

	// Get latest release of the repository
	release, err := g.GitHub.GetLatestRelease(repo)

	if err != nil {
		return nil, err
	}

	// Extract tag
	tag := *release.TagName

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

	return &LatestRelease{tag: tag, hash: hash}, nil
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








