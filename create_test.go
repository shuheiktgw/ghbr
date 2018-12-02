package main

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestCreate(t *testing.T) {
	generator, client, outStream, mux, tearDown := ghbrMockGenerator()
	defer tearDown()

	cmd := NewCreateCmd(generator)
	arg := "ghbr create -t test -o shuheiktgw -r testApp"
	args := strings.Split(arg, " ")
	cmd.SetArgs(args[1:])

	assetPath := fmt.Sprintf("/%s/%s/releases/download/v0.0.1/ghbr_v0.0.1_darwin_amd64.zip", TestOwner, "testApp")
	assetURL := fmt.Sprintf("%s/%s", client.Client.BaseURL, assetPath)

	// Mock GetLatestRelease request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases/latest", TestOwner, "testApp"), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":1,"name":"Release v0.0.1","tag_name":"v0.0.1","assets":[{"name":"ghbr_v0.0.1_darwin_amd64.zip", "browser_download_url":"%s"}]}`, assetURL)
	})

	// Mock downloadFile request
	mux.HandleFunc(assetPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test")
	})

	// Mock CreateRepository request
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBody(t, r, fmt.Sprintf(`{"name":"homebrew-testApp","description":"Homebrew formula for shuheiktgw/testApp","homepage":"https://github.com/shuheiktgw/testApp","private":false}`+"\n"))
		fmt.Fprintf(w, `{"html_url":"https://github.com/shuheiktgw/homebrew-testApp"}`)
	})

	// Mock CreateFile request for README.md
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", TestOwner, "homebrew-testApp", "README.md"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	// Mock CreateFile request for formula file
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", TestOwner, "homebrew-testApp", "testApp.rb"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("#create returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Checking the latest release\n" +
		"[ghbr] ===> Downloading Darwin AMD64 release\n" +
		"[ghbr] ===> Calculating a checksum of the release\n" +
		"[ghbr] ===> Creating a repository\n" +
		"[ghbr] ===> Adding README.md to the repository\n" +
		"[ghbr] ===> Adding testApp.rb to the repository\n" +
		"\n\n" +
		"Yay! Your Homebrew formula repository has been successfully created!\n" +
		"Access https://github.com/shuheiktgw/homebrew-testApp and see what we achieved.\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#create outputed %+v, want %+v", got, expectedOutput)
	}
}
