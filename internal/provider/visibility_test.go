package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mipmip/huphop/internal/config"
)

func TestNormalizeVisibility(t *testing.T) {
	cases := map[string]string{
		"public":   "public",
		"private":  "private",
		"internal": "internal",
		"":         "public",
		"weird":    "public",
		"Public":   "public", // case-sensitive: unknown -> public
	}
	for in, want := range cases {
		if got := NormalizeVisibility(in); got != want {
			t.Errorf("NormalizeVisibility(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestVisibilityFromPrivate(t *testing.T) {
	if got := visibilityFromPrivate(true); got != VisibilityPrivate {
		t.Errorf("private=true -> %q, want private", got)
	}
	if got := visibilityFromPrivate(false); got != VisibilityPublic {
		t.Errorf("private=false -> %q, want public", got)
	}
}

func TestGitHubMapsVisibilityFromPrivate(t *testing.T) {
	t.Setenv("HUPHOP_GITHUB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"name":"pub","owner":{"login":"acme"},"private":false},
			{"name":"priv","owner":{"login":"acme"},"private":true}
		]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_GITHUB_TOKEN"}, Owners: []string{"acme"},
	}
	gh := NewGitHub(cfg, srv.Client())
	repos, err := gh.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	got := visByName(repos)
	if got["pub"] != VisibilityPublic || got["priv"] != VisibilityPrivate {
		t.Fatalf("github visibility = %v", got)
	}
}

func TestCodebergMapsVisibilityFromPrivate(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"name":"pub","owner":{"login":"acme"},"private":false},
			{"name":"priv","owner":{"login":"acme"},"private":true}
		]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}, Owners: []string{"acme"},
	}
	cb := NewCodeberg(cfg, srv.Client())
	repos, err := cb.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	got := visByName(repos)
	if got["pub"] != VisibilityPublic || got["priv"] != VisibilityPrivate {
		t.Fatalf("codeberg visibility = %v", got)
	}
}

func TestGitLabMapsVisibilityValue(t *testing.T) {
	t.Setenv("HUPHOP_GITLAB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/groups/acme/projects", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[
			{"name":"Pub","path":"pub","visibility":"public","namespace":{"full_path":"acme","path":"acme"}},
			{"name":"Priv","path":"priv","visibility":"private","namespace":{"full_path":"acme","path":"acme"}},
			{"name":"Int","path":"int","visibility":"internal","namespace":{"full_path":"acme","path":"acme"}},
			{"name":"None","path":"none","namespace":{"full_path":"acme","path":"acme"}}
		]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "HUPHOP_GITLAB_TOKEN"}, Owners: []string{"acme"},
	}
	gl := NewGitLab(cfg, srv.Client())
	repos, err := gl.ListRepos(context.Background(), cfg.Owners)
	if err != nil {
		t.Fatal(err)
	}
	got := visByName(repos)
	if got["pub"] != VisibilityPublic || got["priv"] != VisibilityPrivate ||
		got["int"] != VisibilityInternal || got["none"] != VisibilityPublic {
		t.Fatalf("gitlab visibility = %v", got)
	}
}

// visByName indexes repo visibility by repo name for concise assertions.
func visByName(repos []Repo) map[string]string {
	out := make(map[string]string, len(repos))
	for _, r := range repos {
		out[r.Name] = r.Visibility
	}
	return out
}
