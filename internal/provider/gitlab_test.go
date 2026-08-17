package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mipmip/skull2/internal/config"
)

func TestGitLabGroupSubgroupsPaginationMapping(t *testing.T) {
	t.Setenv("SKULL2_GITLAB_TOKEN", "gl-tok")

	var gotToken string
	mux := http.NewServeMux()
	mux.HandleFunc("/groups/acme/projects", func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.Header.Get("PRIVATE-TOKEN")
		if r.URL.Query().Get("include_subgroups") != "true" {
			t.Errorf("include_subgroups not set: %q", r.URL.RawQuery)
		}
		switch r.URL.Query().Get("page") {
		case "1":
			w.Header().Set("X-Next-Page", "2")
			fmt.Fprint(w, `[
				{"name":"Alpha","path":"alpha","description":"a","ssh_url_to_repo":"git@gitlab.com:acme/alpha.git",
				 "http_url_to_repo":"https://gitlab.com/acme/alpha.git","web_url":"https://gitlab.com/acme/alpha",
				 "default_branch":"main","archived":false,"last_activity_at":"2022-08-09T10:11:12Z",
				 "namespace":{"full_path":"acme","path":"acme"}}
			]`)
		case "2":
			// no X-Next-Page -> last page. Subgroup project + a fork.
			fmt.Fprint(w, `[
				{"name":"Beta","path":"beta","ssh_url_to_repo":"git@gitlab.com:acme/sub/beta.git",
				 "http_url_to_repo":"https://gitlab.com/acme/sub/beta.git","web_url":"https://gitlab.com/acme/sub/beta",
				 "default_branch":"main","archived":false,
				 "namespace":{"full_path":"acme/sub","path":"sub"}},
				{"name":"Forked","path":"forked","ssh_url_to_repo":"x","http_url_to_repo":"y","web_url":"z",
				 "namespace":{"full_path":"acme","path":"acme"},"forked_from_project":{"id":42}}
			]`)
		default:
			t.Errorf("unexpected page %q", r.URL.Query().Get("page"))
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITLAB_TOKEN"}, Owners: []string{"acme"},
	}
	gl := NewGitLab(cfg, srv.Client())
	repos, err := gl.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	if gotToken != "gl-tok" {
		t.Fatalf("PRIVATE-TOKEN = %q", gotToken)
	}
	if len(repos) != 3 {
		t.Fatalf("got %d repos, want 3: %+v", len(repos), repos)
	}
	alpha := repos[0]
	if alpha.Name != "alpha" || alpha.Owner != "acme" ||
		alpha.SSHURL != "git@gitlab.com:acme/alpha.git" ||
		alpha.HTTPSURL != "https://gitlab.com/acme/alpha.git" ||
		alpha.WebURL != "https://gitlab.com/acme/alpha" || alpha.DefaultBranch != "main" ||
		alpha.Fork {
		t.Fatalf("alpha mapped wrong: %+v", alpha)
	}
	if alpha.UpdatedAt.Year() != 2022 {
		t.Fatalf("last_activity_at not parsed: %v", alpha.UpdatedAt)
	}
	if repos[1].Owner != "acme/sub" {
		t.Fatalf("subgroup namespace not mapped: %+v", repos[1])
	}
	if !repos[2].Fork {
		t.Fatalf("forked_from_project should mark fork: %+v", repos[2])
	}
}

func TestGitLabMembershipWhenNoOwners(t *testing.T) {
	t.Setenv("SKULL2_GITLAB_TOKEN", "gl-tok")
	var hit bool
	mux := http.NewServeMux()
	mux.HandleFunc("/projects", func(w http.ResponseWriter, r *http.Request) {
		hit = true
		if r.URL.Query().Get("membership") != "true" {
			t.Errorf("membership not set: %q", r.URL.RawQuery)
		}
		fmt.Fprint(w, `[{"name":"Own","path":"own","namespace":{"full_path":"me","path":"me"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITLAB_TOKEN"},
	}
	gl := NewGitLab(cfg, srv.Client())
	repos, err := gl.ListRepos(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !hit || len(repos) != 1 || repos[0].Owner != "me" || repos[0].Name != "own" {
		t.Fatalf("membership projects not listed: hit=%v repos=%+v", hit, repos)
	}
}

func TestGitLabFilterArchived(t *testing.T) {
	t.Setenv("SKULL2_GITLAB_TOKEN", "gl-tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/groups/acme/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"name":"A","path":"a","archived":true,"namespace":{"full_path":"acme","path":"acme"}},
			{"name":"B","path":"b","namespace":{"full_path":"acme","path":"acme"}}
		]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITLAB_TOKEN"}, Owners: []string{"acme"},
	}
	gl := NewGitLab(cfg, srv.Client())
	repos, err := gl.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].Name != "b" {
		t.Fatalf("archived not filtered: %+v", repos)
	}
}

func TestGitLabNamespaceFallbackToPath(t *testing.T) {
	t.Setenv("SKULL2_GITLAB_TOKEN", "gl-tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/groups/acme/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"name":"A","path":"a","namespace":{"path":"onlypath"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITLAB_TOKEN"}, Owners: []string{"acme"},
	}
	gl := NewGitLab(cfg, srv.Client())
	repos, err := gl.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 || repos[0].Owner != "onlypath" {
		t.Fatalf("namespace fallback wrong: %+v", repos)
	}
}

func TestGitLabHTTPError(t *testing.T) {
	t.Setenv("SKULL2_GITLAB_TOKEN", "gl-tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer srv.Close()
	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITLAB_TOKEN"}, Owners: []string{"acme"},
	}
	gl := NewGitLab(cfg, srv.Client())
	if _, err := gl.ListRepos(context.Background(), cfg.Owners); err == nil {
		t.Fatal("expected http error")
	}
}

func TestGitLabDefaultBase(t *testing.T) {
	gl := NewGitLab(config.Provider{Type: config.ProviderGitLab}, nil)
	if gl.base != defaultGitLabAPI {
		t.Fatalf("default base = %q", gl.base)
	}
}

func TestNewDefaultRegistryBuildsAllTypes(t *testing.T) {
	reg := NewDefaultRegistry()
	for _, ty := range []config.ProviderType{config.ProviderGitHub, config.ProviderCodeberg, config.ProviderGitLab} {
		p, err := reg.Build(&config.Provider{Type: ty})
		if err != nil {
			t.Fatalf("build %q: %v", ty, err)
		}
		if p == nil {
			t.Fatalf("nil provider for %q", ty)
		}
	}
}
