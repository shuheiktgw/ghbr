package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestGhbr_GetLatestRelease_Success(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}

	assetPath := fmt.Sprintf("/%s/%s/releases/download/v0.0.1/ghbr_v0.0.1_darwin_amd64.zip", TestOwner, TestRepo)
	assetURL := fmt.Sprintf("%s/%s", client.Client.BaseURL, assetPath)

	// Mock GetLatestRelease request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases/latest", TestOwner, TestRepo), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":1,"name":"Release v0.0.1","tag_name":"v0.0.1","assets":[{"name":"ghbr_v0.0.1_darwin_amd64.zip", "browser_download_url":"%s"}]}`, assetURL)
	})

	// Mock downloadFile request
	mux.HandleFunc(assetPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "test")
	})

	got, err := ghbr.GetLatestRelease(TestOwner, TestRepo)
	if err != nil {
		t.Fatalf("#GetLatestRelease returns unexpected error: %s", err)
	}

	expectedRelease := &LatestRelease{version: "v0.0.1", url: assetURL, hash: fmt.Sprintf("%x", sha256.Sum256([]byte("test")))}
	if !reflect.DeepEqual(got, expectedRelease) {
		t.Errorf("#GetLatestRelease returned %+v, want %+v", got, expectedRelease)
	}

	expectedOutput := "[ghbr] ===> Checking the latest release\n" +
		"[ghbr] ===> Downloading Darwin AMD64 release\n" +
		"[ghbr] ===> Calculating a checksum of the release\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#GetLatestRelease outputed %+v, want %+v", got, expectedOutput)
	}
}

func TestGhbr_GetLatestRelease_Fail(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	ghbr := Ghbr{GitHub: client, outStream: ioutil.Discard}

	// Mock GetLatestRelease request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases/latest", TestOwner, TestRepo), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":1,"name":"Release v0.0.1","tag_name":"v0.0.1","assets":[{"name":"ghbr_v0.0.1_darwin_386.zip"}]}`)
	})

	_, err := ghbr.GetLatestRelease(TestOwner, TestRepo)
	if _, ok := err.(*HandledError); !ok {
		t.Fatalf("#GetLatestRelease returns invalid error: %s", err)
	}
}

func TestGhbr_CreateFormula_WithoutOrg(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}

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

	release := LatestRelease{
		version: "v0.0.1",
		url:     "httos://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testAoo_v0.0.1_darwin_amd64.zip",
		hash:    "abcdefg",
	}

	err := ghbr.CreateFormula("", TestOwner, "testApp", "alphabet", false, &release)
	if err != nil {
		t.Fatalf("#CreateFormula returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Creating a repository\n" +
		"[ghbr] ===> Adding README.md to the repository\n" +
		"[ghbr] ===> Adding testApp.rb to the repository\n" +
		"\n\n" +
		"Yay! Your Homebrew formula repository has been successfully created!\n" +
		"Access https://github.com/shuheiktgw/homebrew-testApp and see what we achieved.\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#CreateFormula outputed %+v, want %+v", got, expectedOutput)
	}
}

func TestGhbr_CreateFormula_WithOrg(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}
	org := "TestOrg"

	// Mock CreateRepository request
	mux.HandleFunc("/orgs/TestOrg/repos", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBody(t, r, fmt.Sprintf(`{"name":"homebrew-testApp","description":"Homebrew formula for shuheiktgw/testApp","homepage":"https://github.com/shuheiktgw/testApp","private":false}`+"\n"))
		fmt.Fprintf(w, `{"html_url":"https://github.com/TestOrg/homebrew-testApp"}`)
	})

	// Mock CreateFile request for README.md
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", org, "homebrew-testApp", "README.md"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	// Mock CreateFile request for formula file
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", org, "homebrew-testApp", "testApp.rb"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
	})

	release := LatestRelease{
		version: "v0.0.1",
		url:     "httos://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip",
		hash:    "abcdefg",
	}

	err := ghbr.CreateFormula(org, TestOwner, "testApp", "alphabet", false, &release)
	if err != nil {
		t.Fatalf("#CreateFormula returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Creating a repository\n" +
		"[ghbr] ===> Adding README.md to the repository\n" +
		"[ghbr] ===> Adding testApp.rb to the repository\n" +
		"\n\n" +
		"Yay! Your Homebrew formula repository has been successfully created!\n" +
		"Access https://github.com/TestOrg/homebrew-testApp and see what we achieved.\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#CreateFormula outputed %+v, want %+v", got, expectedOutput)
	}
}

func TestGhbr_UpdateFormulaWithMerge(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}

	content := base64.StdEncoding.EncodeToString([]byte(`
version "v0.0.1"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip"
sha256 "0001123456789012345678901234567890123456789012345678901234567890"
`))

	expectedContent, _ := json.Marshal([]byte(`
version "v0.0.2"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.2/testApp_v0.0.2_darwin_amd64.zip"
sha256 "0002123456789012345678901234567890123456789012345678901234567890"
`))

	// Mock GetFile and UpdateFile request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/homebrew-testApp/contents/testApp.rb", TestOwner), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"path":"testApp.rb","sha":"formulaV0.0.1","encoding":"base64","content":"%s"}`, content)
		case http.MethodPut:
			testBody(t, r, fmt.Sprintf(`{"message":"Bumps up to v0.0.2","content":%s,"sha":"formulaV0.0.1","branch":"bumps_up_to_v0.0.2"}`+"\n", expectedContent))
		}
	})

	release := LatestRelease{
		version: "v0.0.2",
		url:     "https://github.com/shuheiktgw/testApp/releases/download/v0.0.2/testApp_v0.0.2_darwin_amd64.zip",
		hash:    "0002123456789012345678901234567890123456789012345678901234567890",
	}

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
		fmt.Fprintf(w, `{"number":100}`)
	})

	// Mock MergePullRequest request
	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls/%d/merge", TestOwner, "homebrew-testApp", 100), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"sha":"abcdefg","merged":true}`)
	})

	// Mock DeleteLatestRef request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs/%s", TestOwner, "homebrew-testApp", "heads/bumps_up_to_v0.0.2"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	err := ghbr.UpdateFormula("", TestOwner, "testApp", "master", false, true, &release)

	if err != nil {
		t.Fatalf("#UpdateFormula returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Checking the current formula\n" +
		"[ghbr] ===> Creating a new feature branch\n" +
		"[ghbr] ===> Updating the formula file\n" +
		"[ghbr] ===> Creating a Pull Request\n" +
		"[ghbr] ===> Merging the Pull Request\n" +
		"[ghbr] ===> Deleting the branch\n" +
		"\n\n" +
		"Yay! Now your formula is up-to-date!\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#UpdateFormula outputed %+v, want %+v", got, expectedOutput)
	}
}

func TestGhbr_UpdateFormulaWithoutMerge(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}

	content := base64.StdEncoding.EncodeToString([]byte(`
version "v0.0.1"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip"
sha256 "0001123456789012345678901234567890123456789012345678901234567890"
`))

	expectedContent, _ := json.Marshal([]byte(`
version "v0.0.2"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.2/testApp_v0.0.2_darwin_amd64.zip"
sha256 "0002123456789012345678901234567890123456789012345678901234567890"
`))

	// Mock GetFile and UpdateFile request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/homebrew-testApp/contents/testApp.rb", TestOwner), func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			fmt.Fprintf(w, `{"path":"testApp.rb","sha":"formulaV0.0.1","encoding":"base64","content":"%s"}`, content)
		case http.MethodPut:
			testBody(t, r, fmt.Sprintf(`{"message":"Bumps up to v0.0.2","content":%s,"sha":"formulaV0.0.1","branch":"bumps_up_to_v0.0.2"}`+"\n", expectedContent))
		}
	})

	release := LatestRelease{
		version: "v0.0.2",
		url:     "https://github.com/shuheiktgw/testApp/releases/download/v0.0.2/testApp_v0.0.2_darwin_amd64.zip",
		hash:    "0002123456789012345678901234567890123456789012345678901234567890",
	}

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

	err := ghbr.UpdateFormula("", TestOwner, "testApp", "master", false, false, &release)

	if err != nil {
		t.Fatalf("#UpdateFormula returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Checking the current formula\n" +
		"[ghbr] ===> Creating a new feature branch\n" +
		"[ghbr] ===> Updating the formula file\n" +
		"[ghbr] ===> Creating a Pull Request\n" +
		"\n\n" +
		"Yay! Now your formula is ready to update!\n\n" +
		"Access https://github.com/shuheiktgw/homebrew-testApp/pullls/100 and merge the Pull Request\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#UpdateFormula outputed %+v, want %+v", got, expectedOutput)
	}
}

func TestGhbr_UpdateFormula_AlreadyLatest(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}

	content := base64.StdEncoding.EncodeToString([]byte(`
version "v0.0.1"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip"
sha256 "0001123456789012345678901234567890123456789012345678901234567890"
`))

	mux.HandleFunc(fmt.Sprintf("/repos/%s/homebrew-testApp/contents/testApp.rb", TestOwner), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testFormValues(t, r, values{"ref": "master"})
		fmt.Fprintf(w, `{"path":"testApp.rb","encoding":"base64","content":"%s"}`, content)
	})

	release := LatestRelease{
		version: "v0.0.1",
		url:     "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip",
		hash:    "0001123456789012345678901234567890123456789012345678901234567890",
	}

	err := ghbr.UpdateFormula("", TestOwner, "testApp", "master", false, true, &release)

	if err != nil {
		t.Fatalf("#UpdateFormula returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Checking the current formula\n" +
		"\n\n" +
		"ghbr aborted!\n\n" +
		"The current formula (pointing to version v0.0.1) is up-to-date.\n" +
		"If you want to update the formula anyway, run `ghbr release` with `--force` option.\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#UpdateFormula outputed %+v, want %+v", got, expectedOutput)
	}
}

func TestGhbr_UpdateFormulaForceUpdate(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	outStream := new(bytes.Buffer)
	ghbr := Ghbr{GitHub: client, outStream: outStream}

	content := base64.StdEncoding.EncodeToString([]byte(`
version "v0.0.1"
url "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip"
sha256 "0001123456789012345678901234567890123456789012345678901234567890"
`))

	expectedContent, _ := json.Marshal([]byte(`
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
			testBody(t, r, fmt.Sprintf(`{"message":"Bumps up to v0.0.1","content":%s,"sha":"formulaV0.0.1","branch":"bumps_up_to_v0.0.1"}`+"\n", expectedContent))
		}
	})

	release := LatestRelease{
		version: "v0.0.1",
		url:     "https://github.com/shuheiktgw/testApp/releases/download/v0.0.1/testApp_v0.0.1_darwin_amd64.zip",
		hash:    "0001123456789012345678901234567890123456789012345678901234567890",
	}

	// Mock CreateBranch request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs/%s", TestOwner, "homebrew-testApp", "heads/master"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"object":{"sha":"abcdefg"}}`)
	})

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs", TestOwner, "homebrew-testApp"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBody(t, r, fmt.Sprintf(`{"ref":"refs/heads/bumps_up_to_v0.0.1","sha":"abcdefg"}`+"\n"))
		fmt.Fprintf(w, `{"object":{"sha":"abcdefg"}}`)
	})

	// Mock CreatePullRequest request
	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls", TestOwner, "homebrew-testApp"), func(w http.ResponseWriter, r *http.Request) {
		testBody(t, r, fmt.Sprintf(`{"title":"%s","head":"%s","base":"%s","body":"%s"}`+"\n", "Bumps up to v0.0.1", "bumps_up_to_v0.0.1", "master", "Bumps up to v0.0.1"))
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{"number":100}`)
	})

	// Mock MergePullRequest request
	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls/%d/merge", TestOwner, "homebrew-testApp", 100), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"sha":"abcdefg","merged":true}`)
	})

	// Mock DeleteLatestRef request
	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs/%s", TestOwner, "homebrew-testApp", "heads/bumps_up_to_v0.0.1"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	err := ghbr.UpdateFormula("", TestOwner, "testApp", "master", true, true, &release)

	if err != nil {
		t.Fatalf("#UpdateFormula returns unexpected error: %s", err)
	}

	expectedOutput := "[ghbr] ===> Checking the current formula\n" +
		"[ghbr] ===> Creating a new feature branch\n" +
		"[ghbr] ===> Updating the formula file\n" +
		"[ghbr] ===> Creating a Pull Request\n" +
		"[ghbr] ===> Merging the Pull Request\n" +
		"[ghbr] ===> Deleting the branch\n" +
		"\n\n" +
		"Yay! Now your formula is up-to-date!\n\n"

	if got := outStream.String(); got != expectedOutput {
		t.Errorf("#UpdateFormula outputed %+v, want %+v", got, expectedOutput)
	}
}
