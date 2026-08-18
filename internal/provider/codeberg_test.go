package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mipmip/huphop/internal/config"
)

func TestCodebergListReposPaginationAuthMapping(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")

	// Build a full page (codebergPageLimit items) then a short final page.
	fullPage := func() string {
		s := "["
		for i := 0; i < codebergPageLimit; i++ {
			if i > 0 {
				s += ","
			}
			s += fmt.Sprintf(`{"name":"r%d","owner":{"login":"org"},"archived":false,"fork":false}`, i)
		}
		return s + "]"
	}

	var gotAuth string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/org/repos", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if lim := r.URL.Query().Get("limit"); lim != strconv.Itoa(codebergPageLimit) {
			t.Errorf("limit = %q", lim)
		}
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("X-Total-Count", strconv.Itoa(codebergPageLimit+1))
			fmt.Fprint(w, fullPage())
		case "2":
			fmt.Fprint(w, `[{"name":"last","description":"d","owner":{"login":"org"},
				"ssh_url":"git@codeberg.org:org/last.git","clone_url":"https://codeberg.org/org/last.git",
				"html_url":"https://codeberg.org/org/last","default_branch":"main",
				"archived":false,"fork":false,"updated_at":"2024-05-06T07:08:09Z"}]`)
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}, Owners: []string{"org"},
	}
	cb := NewCodeberg(cfg, srv.Client())
	repos, err := cb.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	if gotAuth != "token cb-tok" {
		t.Fatalf("auth = %q, want token cb-tok", gotAuth)
	}
	if len(repos) != codebergPageLimit+1 {
		t.Fatalf("got %d repos, want %d", len(repos), codebergPageLimit+1)
	}
	last := repos[len(repos)-1]
	if last.Name != "last" || last.SSHURL != "git@codeberg.org:org/last.git" ||
		last.HTTPSURL != "https://codeberg.org/org/last.git" ||
		last.WebURL != "https://codeberg.org/org/last" || last.DefaultBranch != "main" {
		t.Fatalf("last mapped wrong: %+v", last)
	}
	if last.UpdatedAt.Year() != 2024 {
		t.Fatalf("updated_at not parsed: %v", last.UpdatedAt)
	}
}

func TestCodebergShortPageStops(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")
	calls := 0
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/org/repos", func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, `[{"name":"a","owner":{"login":"org"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}, Owners: []string{"org"},
	}
	cb := NewCodeberg(cfg, srv.Client())
	repos, err := cb.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 || len(repos) != 1 {
		t.Fatalf("expected single short page: calls=%d repos=%d", calls, len(repos))
	}
}

func TestCodebergFilterArchivedForks(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/org/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"name":"arch","owner":{"login":"org"},"archived":true},
			{"name":"fork","owner":{"login":"org"},"fork":true},
			{"name":"plain","owner":{"login":"org"}}
		]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	inclForks := false
	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}, Owners: []string{"org"},
		IncludeForks: &inclForks,
	}
	cb := NewCodeberg(cfg, srv.Client())
	repos, err := cb.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].Name != "plain" {
		t.Fatalf("filter wrong: %+v", repos)
	}
}

func TestCodebergUserReposWhenNoOwners(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")
	var hit bool
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/repos", func(w http.ResponseWriter, r *http.Request) {
		hit = true
		fmt.Fprint(w, `[{"name":"mine","owner":{"login":"me"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"},
	}
	cb := NewCodeberg(cfg, srv.Client())
	repos, err := cb.ListRepos(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hit || len(repos) != 1 {
		t.Fatalf("user repos not listed: hit=%v repos=%+v", hit, repos)
	}
}

func TestCodebergHTTPError(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusUnauthorized)
	}))
	defer srv.Close()
	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}, Owners: []string{"org"},
	}
	cb := NewCodeberg(cfg, srv.Client())
	if _, err := cb.ListRepos(context.Background(), cfg.Owners); err == nil {
		t.Fatal("expected http error")
	}
}

func TestCodebergDefaultBase(t *testing.T) {
	cb := NewCodeberg(config.Provider{Type: config.ProviderCodeberg}, nil)
	if cb.base != defaultCodebergAPI {
		t.Fatalf("default base = %q", cb.base)
	}
}

func TestCodebergListOwners(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("X-Total-Count", strconv.Itoa(codebergPageLimit+1))
			s := "["
			for i := 0; i < codebergPageLimit; i++ {
				if i > 0 {
					s += ","
				}
				s += fmt.Sprintf(`{"username":"org%d"}`, i)
			}
			fmt.Fprint(w, s+"]")
		case "2":
			fmt.Fprint(w, `[{"username":"last-org"}]`)
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}}
	owners, err := NewCodeberg(cfg, srv.Client()).ListOwners(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(owners) != codebergPageLimit+1 || owners[0] != "org0" || owners[len(owners)-1] != "last-org" {
		t.Fatalf("owners len=%d last=%q", len(owners), owners[len(owners)-1])
	}
}

func TestCodebergListOwnersError(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "cb-tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	cfg := config.Provider{Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}}
	if _, err := NewCodeberg(cfg, srv.Client()).ListOwners(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}
