package main

import (
	"io"
	"os"
	"net/http"
	"github.com/pkg/errors"
)

// GHBR define functions for Homebrew Formula
type GHBR struct {
	GitHub GitHub

	outStream io.Writer
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








