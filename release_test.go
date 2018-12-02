package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestRelease(t *testing.T) {
	generator, client, outStream, mux, tearDown := ghbrMockGenerator()
	defer tearDown()

	cmd := NewReleaseCmd(generator)
	arg := "ghbr release -t test -o shuheiktgw -r testApp"
	args := strings.Split(arg, " ")
	cmd.SetArgs(args[1:])

	assetPath := fmt.Sprintf("/%s/%s/releases/download/v0.0.2/ghbr_v0.0.2_darwin_amd64.zip", TestOwner, "testApp")
	assetURL := fmt.Sprintf("%s/%s", client.Client.BaseURL, assetPath)

	// Mock GetLatestRelease request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases/latest", TestOwner, "testApp"), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":1,"name":"Release v0.0.2","tag_name":"v0.0.2","assets":[{"name":"ghbr_v0.0.2_darwin_amd64.zip", "browser_download_url":"%s"}]}`, assetURL)
	})

	// Mock downloadFile request
	mux.HandleFunc(assetPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test")
	})

	content := base64.StdEncoding.EncodeToString([]byte(`
version "v0.0.1"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip"
sha256 "0001123456789012345678901234567890123456789012345678901234567890"
`))

	// Mock GetFile and UpdateFile request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/homebrew-testApp/contents/testApp.rb", TestOwner), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"path":"testApp.rb","sha":"formulaV0.0.1","encoding":"base64","content":"%s"}`, content)
		case http.MethodPut:
		}
	})

	// Mock CreateBranch request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs/%s", TestOwner, "homebrew-testApp", "heads/master"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"object":{"sha":"abcdefg"}}`)
	})

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs", TestOwner, "homebrew-testApp"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBody(t, r, fmt.Sprintf(`{"ref":"refs/heads/bumps_up_to_v0.0.2","sha":"abcdefg"}`+"\n"))
		fmt.Fprintf(w, `{"object":{"sha":"abcdefg"}}`)
	})

	// Mock CreatePullRequest request
	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls", TestOwner, "homebrew-testApp"), func(w http.ResponseWriter, r *http.Request) {
		testBody(t, r, fmt.Sprintf(`{"title":"%s","head":"%s","base":"%s","body":"%s"}`+"\n", "Bumps up to v0.0.2", "bumps_up_to_v0.0.2", "master", "Bumps up to v0.0.2"))
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{"number":100, "html_url":"https://github.com/shuheiktgw/homebrew-testApp/pullls/100"}`)
	})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("#release returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Checking the latest release\n" +
		"[ghbr] ===> Downloading Darwin AMD64 release\n" +
		"[ghbr] ===> Calculating a checksum of the release\n" +
		"[ghbr] ===> Checking the current formula\n" +
		"[ghbr] ===> Creating a new feature branch\n" +
		"[ghbr] ===> Updating the formula file\n" +
		"[ghbr] ===> Creating a Pull Request\n" +
		"\n\n" +
		"Yay! Now your formula is ready to update!\n\n" +
		"Access https://github.com/shuheiktgw/homebrew-testApp/pullls/100 and merge the Pull Request\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#create outputed %+v, want %+v", got, expectedOutput)
	}
}
