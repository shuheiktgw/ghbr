package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shuheiktgw/ghbr/mocks"
	"github.com/google/go-github/github"
	"encoding/base64"
	"fmt"
)

func TestGHBRClient_GetLatestRelease_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	mockAssets := []github.ReleaseAsset{
		{
			Name:github.String("gemer_v0.0.1_darwin_386.zip"),
			BrowserDownloadURL: github.String("https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_386.zip")},
		{
			Name:github.String("gemer_v0.0.1_darwin_amd64.zip"),
			BrowserDownloadURL: github.String("https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_amd64.zip")},
	}
	mockRelease := github.RepositoryRelease{
		TagName: github.String("v0.0.1"),
		Assets: mockAssets,
	}
	mockGitHub.EXPECT().GetLatestRelease("testRepo").Return(&mockRelease, nil).Times(1)

	ghbr := GHBRClient{GitHub: mockGitHub, outStream: ioutil.Discard}

	lr, err := ghbr.GetCurrentRelease("testRepo")

	if err != nil {
		t.Fatalf("GetCurrentRelease: unexpected error occured: %s", err)
	}

	if got, want := lr.version, "v0.0.1"; got != want {
		t.Fatalf("GetCurrentRelease: wrong version name is returned: got %s, want %s", got, want)
	}
}

func TestGHBRClient_GetLatestRelease_Fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	ghbr := GHBRClient{GitHub: mockGitHub, outStream: ioutil.Discard}

	if _, err := ghbr.GetCurrentRelease(""); err == nil {
		t.Fatalf("GetCurrentRelease: error is not supposed to be nil")
	}
}

func TestGHBRClient_UpdateFormula_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Create mocks
	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	mockContent := base64.StdEncoding.EncodeToString([]byte("version 'v0.0.1' url 'https://github.com' sha256 'c59729923f23bdf90505283f92ae6ac81f90d94ec6a9df916b41daa843590f31'"))
	mockRepositoryContent := github.RepositoryContent{
		Encoding: github.String("base64"),
		Content: github.String(mockContent),
		SHA: github.String("hash"),
	}

	app := "gemer"
	repo := fmt.Sprintf("homebrew-%s", app)
	branch := "master"
	path := fmt.Sprintf("%s.rb", app)
	newVersion := "v0.0.2"
	newSha := "new729923f23bdf90505283f92ae6ac81f90d94ec6a9df916b41daa843590f31"
	newURL := "https://github.com/new"
	newBranch := fmt.Sprintf("bumps_up_to_%s", newVersion)
	newContent := fmt.Sprintf("version '%s' url '%s' sha256 '%s'", newVersion, newURL, newSha)
	message := fmt.Sprintf("Bumps up to %s", newVersion)

	// Stub methods
	mockGitHub.EXPECT().GetFile(repo, branch, path).Return(&mockRepositoryContent, nil).Times(1)
	mockGitHub.EXPECT().CreateBranch(repo, branch, newBranch).Return(nil).Times(1)
	mockGitHub.EXPECT().UpdateFile(repo, newBranch, path, "hash", message, []byte(newContent)).Return(nil).Times(1)
	mockGitHub.EXPECT().CreatePullRequest(repo, message, newBranch, branch, message).Return(nil, nil).Times(1)

	ghbr := GHBRClient{GitHub: mockGitHub, outStream: ioutil.Discard}

	err := ghbr.UpdateFormula(app, branch, &LatestRelease{version: newVersion, url: newURL, hash: newSha})

	if err != nil {
		t.Fatalf("UpdateFormula: unexpected error occured: %s", err)
	}
}

func TestGHBRClient_UpdateFormula_Fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	ghbr := GHBRClient{GitHub: mockGitHub, outStream: ioutil.Discard}

	cases := []struct {
		app, branch string
		release *LatestRelease
	}{
		{app: "", branch: "", release: nil},
		{app: "", branch: "master", release: &LatestRelease{version: "v0.0.1", url: "https://github.com", hash: "123"}},
		{app: "gemer", branch: "", release: &LatestRelease{version: "v0.0.1", url: "https://github.com", hash: "123"}},
		{app: "gemer", branch: "", release: nil},
	}

	for i, tc := range cases {
		if err := ghbr.UpdateFormula(tc.app, tc.branch, tc.release); err == nil {
			t.Fatalf("#%d UpdateFormula: error is not supposed to be nil", i)
		}
	}
}

func TestGHBRClient_DownloadFile_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	ghbr := GHBRClient{GitHub: mockGitHub, outStream: ioutil.Discard}

	path := "gemer_v0.0.1_darwin_amd64.zip"
	url := "https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_amd64.zip"

	err := ghbr.DownloadFile(path, url)
	defer os.Remove(path)

	if err != nil {
		t.Fatalf("DownloadFile: unexpected error occurred: %s", err)
	}
}

func TestGHBRClient_DownloadFile_Fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	ghbr := GHBRClient{GitHub: mockGitHub, outStream: ioutil.Discard}

	cases := []struct {
		path, url string
	}{
		{path: "", url: ""},
		{path: "gemer_v0.0.1_darwin_amd64.zip", url: ""},
		{path: "", url: "https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_amd64.zip"},
	}

	for i, tc := range cases {
		if err := ghbr.DownloadFile(tc.path, tc.url); err == nil {
			t.Fatalf("#%d DownloadFile: error is not supposed to be nil", i)
		}
	}
}
