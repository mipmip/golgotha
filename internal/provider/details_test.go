package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mipmip/huphop/internal/config"
)

func TestGitHubRepoDetailsAndReadme(t *testing.T) {
	t.Setenv("HUPHOP_GITHUB_TOKEN", "tok")
	readme := "# Alpha\n\nHello **world**.\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/acme/alpha", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"stargazers_count":42,"language":"Go","topics":["cli","git"]}`)
	})
	mux.HandleFunc("/repos/acme/alpha/readme", func(w http.ResponseWriter, r *http.Request) {
		enc := base64.StdEncoding.EncodeToString([]byte(readme))
		// GitHub wraps content; a plain string still decodes.
		fmt.Fprintf(w, `{"content":%q,"encoding":"base64"}`, enc)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_GITHUB_TOKEN"}}
	gh := NewGitHub(cfg, srv.Client())

	d, err := gh.RepoDetails(context.Background(), "acme", "alpha")
	if err != nil {
		t.Fatal(err)
	}
	if d.Stars != 42 || d.Language != "Go" || len(d.Topics) != 2 || d.Topics[0] != "cli" {
		t.Fatalf("details mapped wrong: %+v", d)
	}

	md, err := gh.Readme(context.Background(), "acme", "alpha")
	if err != nil {
		t.Fatal(err)
	}
	if md != readme {
		t.Fatalf("readme = %q, want %q", md, readme)
	}
}

func TestGitHubReadmeNotFound(t *testing.T) {
	t.Setenv("HUPHOP_GITHUB_TOKEN", "tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()
	cfg := config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_GITHUB_TOKEN"}}
	gh := NewGitHub(cfg, srv.Client())
	md, err := gh.Readme(context.Background(), "acme", "none")
	if err != nil {
		t.Fatalf("not-found readme should be clean empty, got err %v", err)
	}
	if md != "" {
		t.Fatalf("expected empty readme, got %q", md)
	}
}

func TestGitHubRepoDetailsHTTPError(t *testing.T) {
	t.Setenv("HUPHOP_GITHUB_TOKEN", "tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	cfg := config.Provider{Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_GITHUB_TOKEN"}}
	gh := NewGitHub(cfg, srv.Client())
	if _, err := gh.RepoDetails(context.Background(), "acme", "alpha"); err == nil {
		t.Fatal("expected http error")
	}
}

func TestCodebergRepoDetailsAndReadme(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "tok")
	readme := "# Notes\n\nraw markdown\n"
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/repos/pim/notes", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"stars_count":7,"language":"Markdown"}`)
	})
	mux.HandleFunc("/api/v1/repos/pim/notes/topics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"topics":["docs"]}`)
	})
	mux.HandleFunc("/api/v1/repos/pim/notes/raw/README.md", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, readme)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}}
	cb := NewCodeberg(cfg, srv.Client())

	d, err := cb.RepoDetails(context.Background(), "pim", "notes")
	if err != nil {
		t.Fatal(err)
	}
	if d.Stars != 7 || d.Language != "Markdown" || len(d.Topics) != 1 || d.Topics[0] != "docs" {
		t.Fatalf("details mapped wrong: %+v", d)
	}
	md, err := cb.Readme(context.Background(), "pim", "notes")
	if err != nil {
		t.Fatal(err)
	}
	if md != readme {
		t.Fatalf("readme = %q, want %q", md, readme)
	}
}

func TestCodebergReadmeNotFound(t *testing.T) {
	t.Setenv("HUPHOP_CODEBERG_TOKEN", "tok")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()
	cfg := config.Provider{Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_CODEBERG_TOKEN"}}
	cb := NewCodeberg(cfg, srv.Client())
	md, err := cb.Readme(context.Background(), "pim", "none")
	if err != nil || md != "" {
		t.Fatalf("expected clean empty readme, got md=%q err=%v", md, err)
	}
}

func TestGitLabRepoDetailsAndReadme(t *testing.T) {
	t.Setenv("HUPHOP_GITLAB_TOKEN", "tok")
	readme := "# Widget\n\nGitLab raw\n"
	mux := http.NewServeMux()
	// project details (used for both details and README default-branch lookup).
	mux.HandleFunc("/projects/acme%2Fwidget", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"star_count":5,"topics":["ci"],"default_branch":"main"}`)
	})
	mux.HandleFunc("/projects/acme%2Fwidget/languages", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"Go":80.0,"Shell":20.0}`)
	})
	mux.HandleFunc("/projects/acme%2Fwidget/repository/files/README.md/raw", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ref") != "main" {
			t.Errorf("ref = %q, want main", r.URL.Query().Get("ref"))
		}
		fmt.Fprint(w, readme)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_GITLAB_TOKEN"}}
	gl := NewGitLab(cfg, srv.Client())

	d, err := gl.RepoDetails(context.Background(), "acme", "widget")
	if err != nil {
		t.Fatal(err)
	}
	if d.Stars != 5 || d.Language != "Go" || len(d.Topics) != 1 || d.Topics[0] != "ci" {
		t.Fatalf("details mapped wrong: %+v", d)
	}
	md, err := gl.Readme(context.Background(), "acme", "widget")
	if err != nil {
		t.Fatal(err)
	}
	if md != readme {
		t.Fatalf("readme = %q, want %q", md, readme)
	}
}

func TestGitLabReadmeNotFound(t *testing.T) {
	t.Setenv("HUPHOP_GITLAB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/projects/acme%2Fwidget", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"star_count":0,"default_branch":"main"}`)
	})
	mux.HandleFunc("/projects/acme%2Fwidget/repository/files/README.md/raw", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL, Auth: config.Auth{Env: "HUPHOP_GITLAB_TOKEN"}}
	gl := NewGitLab(cfg, srv.Client())
	md, err := gl.Readme(context.Background(), "acme", "widget")
	if err != nil || md != "" {
		t.Fatalf("expected clean empty readme, got md=%q err=%v", md, err)
	}
}

func TestTopLanguage(t *testing.T) {
	cases := []struct {
		body string
		want string
	}{
		{`{"Go":80.0,"Shell":20.0}`, "Go"},
		{`{"Rust":100}`, "Rust"},
		{`{}`, ""},
		{`not json`, ""},
		// tie broken alphabetically for determinism
		{`{"Zig":50,"Ada":50}`, "Ada"},
	}
	for _, c := range cases {
		if got := topLanguage([]byte(c.body)); got != c.want {
			t.Errorf("topLanguage(%s) = %q, want %q", c.body, got, c.want)
		}
	}
}
