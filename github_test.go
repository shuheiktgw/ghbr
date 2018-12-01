package main

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-github/github"
)

const (
	TestOwner = "shuheiktgw"
	TestRepo  = "ghbr"
)

func TestNewGitHubClient(t *testing.T) {
	c := NewGitHubClient("test")
	if c == nil {
		t.Fatalf("#NewGitHubClient returns empty GitHubClient")
	}
}

func TestGitHubClient_GetLatestRelease(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/releases/latest", TestOwner, TestRepo), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"id":1,"name":"Release v0.0.1","Draft":false}`)
	})

	rr, err := client.GetLatestRelease(TestOwner, TestRepo)
	if err != nil {
		t.Fatalf("#GetLatestRelease returns unexpected error: %v", err)
	}

	release := github.RepositoryRelease{ID: github.Int64(1), Name: github.String("Release v0.0.1"), Draft: github.Bool(false)}
	if !reflect.DeepEqual(rr, &release) {
		t.Errorf("#GetLatestRelease returned %+v, want %+v", rr, release)
	}
}

func TestGitHubClient_CreateBranch(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	originBranch := "master"
	newBranch := "develop"
	sha := "abcdefgh"

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs/%s", TestOwner, TestRepo, "heads/"+originBranch), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"object":{"sha":"%s"}}`, sha)
	})

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs", TestOwner, TestRepo), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBody(t, r, fmt.Sprintf(`{"ref":"refs/heads/%s","sha":"%s"}`+"\n", newBranch, sha))
		fmt.Fprintf(w, `{"object":{"sha":"%s"}}`, sha)
	})

	err := client.CreateBranch(TestOwner, TestRepo, originBranch, newBranch)
	if err != nil {
		t.Fatalf("#CreateBranch returns unexpected error: %v", err)
	}
}

func TestGitHubClient_DeleteLatestRef(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	branch := "develop"

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/git/refs/%s", TestOwner, TestRepo, "heads/"+branch), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	err := client.DeleteLatestRef(TestOwner, TestRepo, branch)
	if err != nil {
		t.Fatalf("#DeleteLatestRef returns unexpected error: %v", err)
	}
}

func TestGitHubClient_CreatePullRequest(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	title := "Test PR"
	head := "develop"
	base := "feature"
	body := "This is a PR for test"

	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls", TestOwner, TestRepo), func(w http.ResponseWriter, r *http.Request) {
		testBody(t, r, fmt.Sprintf(`{"title":"%s","head":"%s","base":"%s","body":"%s"}`+"\n", title, head, base, body))
		testMethod(t, r, http.MethodPost)
		fmt.Fprintf(w, `{"title":"%s","body":"%s"}`, title, body)
	})

	pr, err := client.CreatePullRequest(TestOwner, TestRepo, title, head, base, body)
	if err != nil {
		t.Fatalf("#CreatePullRequest returns unexpected error: %v", err)
	}

	want := github.PullRequest{Title: github.String(title), Body: github.String(body)}
	if !reflect.DeepEqual(pr, &want) {
		t.Errorf("#CreatePullRequest returned %+v, want %+v", pr, &want)
	}
}

func TestGitHubClient_MergePullRequest(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	number := 1

	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls/%d/merge", TestOwner, TestRepo, number), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"sha":"abcdefg","merged":true}`)
	})

	err := client.MergePullRequest(TestOwner, TestRepo, number)
	if err != nil {
		t.Fatalf("#MergePullRequest returns unexpected error: %v", err)
	}
}

func TestGitHubClient_ClosePullRequest(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	number := 1

	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v/pulls/%d", TestOwner, TestRepo, number), func(w http.ResponseWriter, r *http.Request) {
		testBody(t, r, `{"state":"close"}`+"\n")
		testMethod(t, r, http.MethodPatch)
		fmt.Fprint(w, `{"number":1}`)
	})

	err := client.ClosePullRequest(TestOwner, TestRepo, number)
	if err != nil {
		t.Fatalf("#ClosePullRequest returns unexpected error: %v", err)
	}
}

func TestGitHubClient_GetFile(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	path := "test"
	branch := "develop"

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", TestOwner, TestRepo, path), func(w http.ResponseWriter, r *http.Request) {
		testFormValues(t, r, values{"ref": branch})
		testMethod(t, r, http.MethodGet)
		fmt.Fprintf(w, `{"path":"%s"}`, path)
	})

	rc, err := client.GetFile(TestOwner, TestRepo, branch, path)
	if err != nil {
		t.Fatalf("#GetFile returns unexpected error: %v", err)
	}

	want := &github.RepositoryContent{Path: &path}
	if !reflect.DeepEqual(rc, want) {
		t.Errorf("#GetFile returned %+v, want %+v", rc, want)
	}
}

func TestGitHubClient_CreateFile(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	branch := "develop"
	path := "test"
	message := "This is a test"
	content := "Test"

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", TestOwner, TestRepo, path), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testBody(t, r, fmt.Sprintf(`{"message":"%s","content":"%s","branch":"%s"}`+"\n", message, "VGVzdA==", branch))
		fmt.Fprintf(w, `{"content":{"path":"%s","content":"%s"}}`, path, content)
	})

	rc, err := client.CreateFile(TestOwner, TestRepo, branch, path, message, []byte(content))
	if err != nil {
		t.Fatalf("#CreateFile returns unexpected error: %v", err)
	}

	want := &github.RepositoryContentResponse{Content: &github.RepositoryContent{Path: &path, Content: &content}}
	if !reflect.DeepEqual(rc, want) {
		t.Errorf("#CreateFile returned %+v, want %+v", rc, want)
	}
}

func TestGitHubClient_UpdateFile(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	branch := "develop"
	path := "test"
	sha := "abcdefg"
	message := "This is a test"
	content := "Test"

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", TestOwner, TestRepo, path), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testBody(t, r, fmt.Sprintf(`{"message":"%s","content":"%s","sha":"%s","branch":"%s"}`+"\n", message, "VGVzdA==", sha, branch))
		fmt.Fprintf(w, `{"content":{"path":"%s","content":"%s"}}`, path, content)
	})

	err := client.UpdateFile(TestOwner, TestRepo, branch, path, sha, message, []byte(content))
	if err != nil {
		t.Fatalf("#UpdateFile returns unexpected error: %v", err)
	}
}

func TestGitHubClient_DeleteFile(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	branch := "develop"
	path := "test"
	sha := "abcdefg"
	message := "This is a test"

	mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/contents/%s", TestOwner, TestRepo, path), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		testBody(t, r, fmt.Sprintf(`{"message":"%s","sha":"%s","branch":"%s"}`+"\n", message, sha, branch))
		fmt.Fprintf(w, `{"content":{"path":"%s"}}`, path)
	})

	err := client.DeleteFile(TestOwner, TestRepo, branch, path, sha, message)
	if err != nil {
		t.Fatalf("#DeleteFile returns unexpected error: %v", err)
	}
}

func TestGitHubClient_CreateRepository(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	org := "shuheiktgw"
	name := "testRepo"
	description := "This is a test"
	homepage := "https://shuheiktgw.com"
	private := false

	mux.HandleFunc(fmt.Sprintf("/orgs/%v/repos", org), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testBody(t, r, fmt.Sprintf(`{"name":"%s","description":"%s","homepage":"%s","private":%v}`+"\n", name, description, homepage, private))
		fmt.Fprintf(w, `{"name":"%s","description":"%s","homepage":"%s"}`, name, description, homepage)
	})

	repo, err := client.CreateRepository(org, name, description, homepage, private)
	if err != nil {
		t.Fatalf("#CreateRepository returns unexpected error: %v", err)
	}

	want := &github.Repository{Name: &name, Description: &description, Homepage: &homepage}
	if !reflect.DeepEqual(repo, want) {
		t.Errorf("#CreateRepository returned %+v, want %+v", repo, want)
	}
}

func TestGitHubClient_DeleteRepository(t *testing.T) {
	client, mux, _, tearDown := setup()
	defer tearDown()

	org := "shuheiktgw"
	name := "testRepo"

	mux.HandleFunc(fmt.Sprintf("/repos/%v/%v", org, name), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	err := client.DeleteRepository(org, name)
	if err != nil {
		t.Fatalf("#DeleteRepository returns unexpected error: %v", err)
	}
}
