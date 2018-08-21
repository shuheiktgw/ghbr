package hbr

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	goGitHub "github.com/google/go-github/github"
	"github.com/pkg/errors"
	"github.com/shuheiktgw/ghbr/github"
)

func TestGenerateHBR(t *testing.T) {
	cases := []struct {
		token, owner          string
		clientExist, errExist bool
	}{
		{token: "", owner: "", errExist: true},
		{token: "test", owner: "", errExist: true},
		{token: "", owner: "test", errExist: true},
		{token: "test", owner: "test", errExist: false},
	}

	for i, tc := range cases {
		g := GenerateHBR(tc.token, tc.owner)

		if tc.errExist && g.Err() == nil {
			t.Fatalf("#%d Error is not supposed to be nil", i)
		}

		if !tc.errExist && g.Err() != nil {
			t.Fatalf("#%d Error is supposed to be nil", i)
		}
	}
}

func TestHBRWrapper_GetCurrentRelease(t *testing.T) {
	cases := []struct {
		err   error
		count int
	}{
		{err: nil, count: 1},
		{err: errors.New("error!"), count: 0},
	}

	for _, tc := range cases {
		mockCtrl := gomock.NewController(t)
		mockClient := NewMockHBRClient(mockCtrl)

		mockClient.EXPECT().GetCurrentRelease("testRepo").Return(&LatestRelease{}, nil).Times(tc.count)

		wrapper := Wrapper{client: mockClient, err: tc.err}
		wrapper.GetCurrentRelease("testRepo")

		mockCtrl.Finish()
	}
}

func TestHBRWrapper_CreateFormula(t *testing.T) {
	cases := []struct {
		err   error
		count int
	}{
		{err: nil, count: 1},
		{err: errors.New("error!"), count: 0},
	}

	for _, tc := range cases {
		mockCtrl := gomock.NewController(t)
		mockClient := NewMockHBRClient(mockCtrl)

		release := &LatestRelease{}
		mockClient.EXPECT().CreateFormula("testApp", "test", false, release).Return(nil).Times(tc.count)

		wrapper := Wrapper{client: mockClient, err: tc.err}
		wrapper.CreateFormula("testApp", "test", false, release)

		mockCtrl.Finish()
	}
}

func TestHBRWrapper_UpdateFormula(t *testing.T) {
	cases := []struct {
		err   error
		count int
	}{
		{err: nil, count: 1},
		{err: errors.New("error!"), count: 0},
	}

	for _, tc := range cases {
		mockCtrl := gomock.NewController(t)
		mockClient := NewMockHBRClient(mockCtrl)

		release := &LatestRelease{}
		mockClient.EXPECT().UpdateFormula("testApp", "master", false, false, release).Return(nil).Times(tc.count)

		wrapper := Wrapper{client: mockClient, err: tc.err}
		wrapper.UpdateFormula("testApp", "master", false, false, release)

		mockCtrl.Finish()
	}
}

func TestHBRClient_GetCurrentReleaseSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := github.NewMockGitHub(mockCtrl)
	mockAssets := []goGitHub.ReleaseAsset{
		{
			Name:               goGitHub.String("gemer_v0.0.1_darwin_386.zip"),
			BrowserDownloadURL: goGitHub.String("https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_386.zip")},
		{
			Name:               goGitHub.String("gemer_v0.0.1_darwin_amd64.zip"),
			BrowserDownloadURL: goGitHub.String("https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_amd64.zip")},
	}
	mockRelease := goGitHub.RepositoryRelease{
		TagName: goGitHub.String("v0.0.1"),
		Assets:  mockAssets,
	}
	mockGitHub.EXPECT().GetLatestRelease("testRepo").Return(&mockRelease, nil).Times(1)

	hbr := Client{GitHub: mockGitHub}

	lr, err := hbr.GetCurrentRelease("testRepo")

	if err != nil {
		t.Fatalf("GetCurrentRelease: unexpected error occured: %s", err)
	}

	if got, want := lr.version, "v0.0.1"; got != want {
		t.Fatalf("GetCurrentRelease: wrong version name is returned: got %s, want %s", got, want)
	}
}

func TestHBRClient_CreateFormulaFail(t *testing.T) {
	cases := []struct {
		app, font string
		release   *LatestRelease
	}{
		{app: "", font: "", release: nil},
		{app: "", font: "test", release: &LatestRelease{}},
		{app: "gemer", font: "", release: &LatestRelease{}},
		{app: "gemer", font: "test", release: nil},
	}

	for i, tc := range cases {
		mockCtrl := gomock.NewController(t)
		mockGitHub := github.NewMockGitHub(mockCtrl)
		hbr := Client{GitHub: mockGitHub}

		if err := hbr.CreateFormula(tc.app, tc.font, false, tc.release); err == nil {
			t.Fatalf("#%d CreateFormula: error is not supposed to be nil", i)
		}

		mockCtrl.Finish()
	}
}

func TestHBRClient_CreateFormulaSuccess(t *testing.T) {
	// Create mock
	mockCtrl := gomock.NewController(t)
	mockGitHub := github.NewMockGitHub(mockCtrl)

	// Stub methods
	mockGitHub.EXPECT().GetOwner().Return("shuheiktgw").Times(1)
	mockGitHub.EXPECT().CreateRepository(
		"homebrew-gemer",
		"Homebrew formula for shuheiktgw/gemer",
		"https://github.com/shuheiktgw/gemer",
		false,
	).Return(&goGitHub.Repository{HTMLURL: goGitHub.String("test.com")}, nil).Times(1)
	mockGitHub.EXPECT().CreateFile(
		"homebrew-gemer",
		"master",
		gomock.Any(),
		gomock.Any(),
		gomock.Any(),
	).Return(nil, nil).Times(2)

	hbr := Client{GitHub: mockGitHub}

	if err := hbr.CreateFormula("gemer", "isometric3", false, &LatestRelease{version: "v0.0.1", url: "test.com", hash: "fdksahuiasfa"}); err != nil {
		t.Fatalf("CreateFormula: unexpected error occured: %s", err)
	}

	mockCtrl.Finish()
}

func TestHBRClient_UpdateFormulaSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Create mocks
	mockGitHub := github.NewMockGitHub(mockCtrl)
	mockContent := base64.StdEncoding.EncodeToString([]byte("version 'v0.0.1' url 'https://github.com' sha256 'c59729923f23bdf90505283f92ae6ac81f90d94ec6a9df916b41daa843590f31'"))
	mockRepositoryContent := goGitHub.RepositoryContent{
		Encoding: goGitHub.String("base64"),
		Content:  goGitHub.String(mockContent),
		SHA:      goGitHub.String("hash"),
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
	mockGitHub.EXPECT().CreatePullRequest(repo, message, newBranch, branch, message).Return(&goGitHub.PullRequest{HTMLURL: goGitHub.String("test.com")}, nil).Times(1)

	hbr := Client{GitHub: mockGitHub}

	err := hbr.UpdateFormula(app, branch, false, false, &LatestRelease{version: newVersion, url: newURL, hash: newSha})

	if err != nil {
		t.Fatalf("UpdateFormula: unexpected error occured: %s", err)
	}
}

func TestHBRClient_UpdateFormulaFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := github.NewMockGitHub(mockCtrl)
	hbr := Client{GitHub: mockGitHub}

	cases := []struct {
		app, branch string
		release     *LatestRelease
	}{
		{app: "", branch: "", release: nil},
		{app: "", branch: "master", release: &LatestRelease{version: "v0.0.1", url: "https://github.com", hash: "123"}},
		{app: "gemer", branch: "", release: &LatestRelease{version: "v0.0.1", url: "https://github.com", hash: "123"}},
		{app: "gemer", branch: "", release: nil},
	}

	for i, tc := range cases {
		if err := hbr.UpdateFormula(tc.app, tc.branch, false, false, tc.release); err == nil {
			t.Fatalf("#%d UpdateFormula: error is not supposed to be nil", i)
		}
	}
}

func TestHBRClient_DownloadFileSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := github.NewMockGitHub(mockCtrl)
	hbr := Client{GitHub: mockGitHub}

	url := "https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_amd64.zip"

	_, err := hbr.downloadFile(url)

	if err != nil {
		t.Fatalf("downloadFile: unexpected error occurred: %s", err)
	}
}

func TestHBRClient_DownloadFileFail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := github.NewMockGitHub(mockCtrl)
	hbr := Client{GitHub: mockGitHub}

	cases := []struct {
		path, url string
	}{
		{url: ""},
	}

	for i, tc := range cases {
		if _, err := hbr.downloadFile(tc.url); err == nil {
			t.Fatalf("#%d downloadFile: error is not supposed to be nil", i)
		}
	}
}
