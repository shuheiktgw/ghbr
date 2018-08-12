package main

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shuheiktgw/ghbr/mocks"
)

func TestGHBR_DownloadFile_Success(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	ghbr := GHBR{GitHub: mockGitHub, outStream: ioutil.Discard}

	path := "gemer_v0.0.1_darwin_amd64.zip"
	url := "https://github.com/shuheiktgw/gemer/releases/download/0.0.1/gemer_v0.0.1_darwin_amd64.zip"

	err := ghbr.DownloadFile(path, url)
	defer os.Remove(path)

	if err != nil {
		t.Fatalf("DownloadFile: unexpected error occurred: %s", err)
	}
}

func TestGHBR_DownloadFile_Fail(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockGitHub := mocks.NewMockGitHub(mockCtrl)
	ghbr := GHBR{GitHub: mockGitHub, outStream: ioutil.Discard}

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

