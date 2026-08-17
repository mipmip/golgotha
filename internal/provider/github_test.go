package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mipmip/golgotha/internal/config"
)

func TestGitHubListReposPaginationAndMapping(t *testing.T) {
	t.Setenv("GOLGOTHA_GITHUB_TOKEN", "tok-123")

	var gotAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		page := r.URL.Query().Get("page")
		switch page {
		case "", "1":
			host := "http://" + r.Host
			w.Header().Set("Link", fmt.Sprintf(`<%s/orgs/acme/repos?per_page=100&page=2>; rel="next", <%s/orgs/acme/repos?per_page=100&page=2>; rel="last"`, host, host))
			fmt.Fprint(w, `[
				{"name":"alpha","description":"first","owner":{"login":"acme"},
				 "ssh_url":"git@github.com:acme/alpha.git","clone_url":"https://github.com/acme/alpha.git",
				 "html_url":"https://github.com/acme/alpha","default_branch":"main",
				 "archived":false,"fork":false,"updated_at":"2023-01-02T03:04:05Z"}
			]`)
		case "2":
			fmt.Fprint(w, `[
				{"name":"beta","description":"","owner":{"login":"acme"},
				 "ssh_url":"git@github.com:acme/beta.git","clone_url":"https://github.com/acme/beta.git",
				 "html_url":"https://github.com/acme/beta","default_branch":"master",
				 "archived":true,"fork":false,"updated_at":"2023-02-02T03:04:05Z"},
				{"name":"gamma","description":"","owner":{"login":"acme"},
				 "ssh_url":"git@github.com:acme/gamma.git","clone_url":"https://github.com/acme/gamma.git",
				 "html_url":"https://github.com/acme/gamma","default_branch":"main",
				 "archived":false,"fork":true,"updated_at":"2023-03-02T03:04:05Z"}
			]`)
		default:
			t.Errorf("unexpected page %q", page)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Default filters: exclude archived, include forks.
	cfg := config.Provider{
		Name:   "github",
		Type:   config.ProviderGitHub,
		Short:  "gh",
		APIURL: srv.URL,
		Auth:   config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN"},
		Owners: []string{"acme"},
	}
	gh := NewGitHub(cfg, srv.Client())
	gh.env = osLookupEnv

	repos, err := gh.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}

	if gotAuth != "Bearer tok-123" {
		t.Fatalf("auth header = %q, want Bearer tok-123", gotAuth)
	}

	// beta is archived (excluded); alpha and gamma (fork included) remain.
	if len(repos) != 2 {
		t.Fatalf("got %d repos, want 2: %+v", len(repos), repos)
	}
	a := repos[0]
	if a.Name != "alpha" || a.Owner != "acme" || a.Description != "first" ||
		a.SSHURL != "git@github.com:acme/alpha.git" ||
		a.HTTPSURL != "https://github.com/acme/alpha.git" ||
		a.WebURL != "https://github.com/acme/alpha" ||
		a.DefaultBranch != "main" || a.Archived || a.Fork {
		t.Fatalf("alpha mapped wrong: %+v", a)
	}
	if a.UpdatedAt.Year() != 2023 {
		t.Fatalf("alpha updated_at not parsed: %v", a.UpdatedAt)
	}
	if repos[1].Name != "gamma" || !repos[1].Fork {
		t.Fatalf("expected gamma fork second: %+v", repos[1])
	}
}

func TestGitHubIncludeArchivedAndExcludeForks(t *testing.T) {
	t.Setenv("GOLGOTHA_GITHUB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"name":"arch","owner":{"login":"acme"},"archived":true,"fork":false},
			{"name":"fork","owner":{"login":"acme"},"archived":false,"fork":true},
			{"name":"plain","owner":{"login":"acme"}}
		]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	inclArch := true
	inclForks := false
	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN"}, Owners: []string{"acme"},
		IncludeArchived: &inclArch, IncludeForks: &inclForks,
	}
	gh := NewGitHub(cfg, srv.Client())
	repos, err := gh.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, r := range repos {
		names[r.Name] = true
	}
	if !names["arch"] || names["fork"] || !names["plain"] {
		t.Fatalf("filter wrong: %v", names)
	}
}

func TestGitHubAuthenticatedUserWhenNoOwners(t *testing.T) {
	t.Setenv("GOLGOTHA_GITHUB_TOKEN", "tok")
	var hitUserRepos bool
	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		hitUserRepos = true
		fmt.Fprint(w, `[{"name":"me","owner":{"login":"pim"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN"},
	}
	gh := NewGitHub(cfg, srv.Client())
	repos, err := gh.ListRepos(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hitUserRepos || len(repos) != 1 || repos[0].Name != "me" {
		t.Fatalf("user repos not listed: hit=%v repos=%+v", hitUserRepos, repos)
	}
}

func TestGitHubAuthErrorNoToken(t *testing.T) {
	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh",
		Auth: config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN_MISSING_XYZ"},
	}
	gh := NewGitHub(cfg, http.DefaultClient)
	gh.getter = stubGetter{}
	gh.env = env(map[string]string{})
	if _, err := gh.ListRepos(context.Background(), []string{"acme"}); err == nil {
		t.Fatal("expected auth error")
	}
}

func TestGitHubHTTPError(t *testing.T) {
	t.Setenv("GOLGOTHA_GITHUB_TOKEN", "tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN"}, Owners: []string{"acme"},
	}
	gh := NewGitHub(cfg, srv.Client())
	if _, err := gh.ListRepos(context.Background(), cfg.Owners); err == nil {
		t.Fatal("expected http error")
	}
}

func TestGitHubDefaultBaseURL(t *testing.T) {
	gh := NewGitHub(config.Provider{Type: config.ProviderGitHub}, nil)
	if gh.base != defaultGitHubAPI {
		t.Fatalf("default base = %q", gh.base)
	}
}

func TestNextLink(t *testing.T) {
	cases := map[string]string{
		"": "",
		`<https://api/x?page=2>; rel="next", <https://api/x?page=5>; rel="last"`: "https://api/x?page=2",
		`<https://api/x?page=5>; rel="last"`:                                     "",
		`<https://api/x?page=2>; rel=next`:                                       "https://api/x?page=2",
		`malformed`:                                                              "",
	}
	for header, want := range cases {
		if got := nextLink(header); got != want {
			t.Errorf("nextLink(%q) = %q, want %q", header, got, want)
		}
	}
}

func TestGitHubListOwnersPagination(t *testing.T) {
	t.Setenv("GOLGOTHA_GITHUB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "", "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/user/orgs?per_page=100&page=2>; rel="next"`, "http://"+r.Host))
			fmt.Fprint(w, `[{"login":"acme"},{"login":"beta"}]`)
		case "2":
			fmt.Fprint(w, `[{"login":"gamma"}]`)
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN"},
	}
	gh := NewGitHub(cfg, srv.Client())
	owners, err := gh.ListOwners(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(owners) != 3 || owners[0] != "acme" || owners[1] != "beta" || owners[2] != "gamma" {
		t.Fatalf("owners = %v, want [acme beta gamma]", owners)
	}
}

func TestGitHubListOwnersEmptyAndError(t *testing.T) {
	t.Setenv("GOLGOTHA_GITHUB_TOKEN", "tok")
	// Empty discovery.
	empty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	defer empty.Close()
	cfg := config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: empty.URL, Auth: config.Auth{Env: "GOLGOTHA_GITHUB_TOKEN"}}
	owners, err := NewGitHub(cfg, empty.Client()).ListOwners(context.Background())
	if err != nil || len(owners) != 0 {
		t.Fatalf("empty discovery: owners=%v err=%v", owners, err)
	}

	// HTTP error.
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer bad.Close()
	cfg.APIURL = bad.URL
	if _, err := NewGitHub(cfg, bad.Client()).ListOwners(context.Background()); err == nil {
		t.Fatal("expected error from forbidden orgs")
	}
}
