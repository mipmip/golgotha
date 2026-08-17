package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/fetch"
)

// evRecorder collects fetch events thread-safely.
type evRecorder struct {
	mu sync.Mutex
	ev []fetch.Event
}

func (r *evRecorder) emit(e fetch.Event) {
	r.mu.Lock()
	r.ev = append(r.ev, e)
	r.mu.Unlock()
}

func (r *evRecorder) count(k fetch.Kind) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, e := range r.ev {
		if e.Kind == k {
			n++
		}
	}
	return n
}

func (r *evRecorder) maxTotalPages() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	m := 0
	for _, e := range r.ev {
		if e.TotalPages > m {
			m = e.TotalPages
		}
	}
	return m
}

// --- GitHub: total from rel="last", fan-out pages 2..N, events emitted. ---

func TestGitHubFetchOwnerDeterminate(t *testing.T) {
	t.Setenv("SKULL2_GITHUB_TOKEN", "tok")
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		host := "http://" + r.Host
		page := r.URL.Query().Get("page")
		switch page {
		case "1":
			w.Header().Set("Link", fmt.Sprintf(`<%s/orgs/acme/repos?page=3>; rel="last"`, host))
			fmt.Fprint(w, `[{"name":"a","owner":{"login":"acme"}}]`)
		case "2":
			fmt.Fprint(w, `[{"name":"b","owner":{"login":"acme"}}]`)
		case "3":
			fmt.Fprint(w, `[{"name":"c","owner":{"login":"acme"}}]`)
		default:
			t.Errorf("unexpected page %q", page)
		}
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITHUB_TOKEN"},
	}
	gh := NewGitHub(cfg, srv.Client())

	rec := &evRecorder{}
	repos, err := gh.FetchOwner(context.Background(), rec.emit, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 3 {
		t.Fatalf("got %d repos, want 3: %+v", len(repos), repos)
	}
	if rec.maxTotalPages() != 3 {
		t.Fatalf("total pages should be 3 from rel=last, got %d", rec.maxTotalPages())
	}
	if rec.count(fetch.KindStarted) != 1 || rec.count(fetch.KindDone) != 1 {
		t.Fatal("expected exactly one Started and one Done")
	}
	if rec.count(fetch.KindPageDone) != 3 {
		t.Fatalf("PageDone = %d, want 3", rec.count(fetch.KindPageDone))
	}
}

func TestGitHubFetchOwnerSelf(t *testing.T) {
	t.Setenv("SKULL2_GITHUB_TOKEN", "tok")
	var hit bool
	mux := http.NewServeMux()
	mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		hit = true
		fmt.Fprint(w, `[{"name":"me","owner":{"login":"pim"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITHUB_TOKEN"},
	}
	repos, err := NewGitHub(cfg, srv.Client()).FetchOwner(context.Background(), nil, config.SelfOwner)
	if err != nil {
		t.Fatal(err)
	}
	if !hit || len(repos) != 1 || repos[0].Name != "me" {
		t.Fatalf("self fetch wrong: hit=%v repos=%+v", hit, repos)
	}
}

func TestGitHubFetchOwnerCancel(t *testing.T) {
	t.Setenv("SKULL2_GITHUB_TOKEN", "tok")
	ctx, cancel := context.WithCancel(context.Background())
	mux := http.NewServeMux()
	mux.HandleFunc("/orgs/acme/repos", func(w http.ResponseWriter, r *http.Request) {
		cancel() // cancel as soon as the first request arrives
		fmt.Fprint(w, `[{"name":"a","owner":{"login":"acme"}}]`)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "github", Type: config.ProviderGitHub, Short: "gh", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITHUB_TOKEN"},
	}
	rec := &evRecorder{}
	_, err := NewGitHub(cfg, srv.Client()).FetchOwner(ctx, rec.emit, "acme")
	if err == nil {
		t.Fatal("expected cancellation error")
	}
	if rec.count(fetch.KindCanceled) != 1 {
		t.Fatal("expected a Canceled event")
	}
}

// --- Codeberg: total derived from X-Total-Count. ---

func TestCodebergFetchOwnerTotalFromHeader(t *testing.T) {
	t.Setenv("SKULL2_CODEBERG_TOKEN", "cb")
	// 3 pages at limit 50 -> X-Total-Count between 101 and 150.
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/orgs/org/repos", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "1" {
			w.Header().Set("X-Total-Count", "120")
		}
		// Each page returns one repo named after the page.
		fmt.Fprintf(w, `[{"name":"p%s","owner":{"login":"org"}}]`, page)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "codeberg", Type: config.ProviderCodeberg, Short: "cb", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_CODEBERG_TOKEN"},
	}
	rec := &evRecorder{}
	repos, err := NewCodeberg(cfg, srv.Client()).FetchOwner(context.Background(), rec.emit, "org")
	if err != nil {
		t.Fatal(err)
	}
	// 120 items / 50 per page => 3 pages fetched; each returns 1 repo here.
	if rec.maxTotalPages() != 3 {
		t.Fatalf("total pages = %d, want 3", rec.maxTotalPages())
	}
	if len(repos) != 3 {
		t.Fatalf("got %d repos, want 3", len(repos))
	}
}

func TestCodebergTotalPages(t *testing.T) {
	cases := map[string]int{
		"":    0,
		"abc": 0,
		"0":   0,
		"1":   1,
		"50":  1,
		"51":  2,
		"100": 2,
		"120": 3,
		"-5":  0,
	}
	for in, want := range cases {
		if got := codebergTotalPages(in); got != want {
			t.Errorf("codebergTotalPages(%q) = %d, want %d", in, got, want)
		}
	}
}

// --- GitLab: total from X-Total-Pages. ---

func TestGitLabFetchOwnerTotalPages(t *testing.T) {
	t.Setenv("SKULL2_GITLAB_TOKEN", "gl")
	mux := http.NewServeMux()
	mux.HandleFunc("/groups/acme/projects", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		if page == "1" {
			w.Header().Set("X-Total-Pages", "2")
		}
		fmt.Fprintf(w, `[{"name":"P%s","path":"p%s","namespace":{"full_path":"acme","path":"acme"}}]`, page, page)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()
	cfg := config.Provider{
		Name: "gitlab", Type: config.ProviderGitLab, Short: "gl", APIURL: srv.URL,
		Auth: config.Auth{Env: "SKULL2_GITLAB_TOKEN"},
	}
	rec := &evRecorder{}
	repos, err := NewGitLab(cfg, srv.Client()).FetchOwner(context.Background(), rec.emit, "acme")
	if err != nil {
		t.Fatal(err)
	}
	if rec.maxTotalPages() != 2 {
		t.Fatalf("total pages = %d, want 2", rec.maxTotalPages())
	}
	if len(repos) != 2 {
		t.Fatalf("got %d repos, want 2", len(repos))
	}
}

// --- lastLinkPage parsing. ---

func TestLastLinkPage(t *testing.T) {
	cases := map[string]int{
		"": 0,
		`<https://api/x?page=2>; rel="next", <https://api/x?page=9>; rel="last"`: 9,
		`<https://api/x?page=2>; rel="next"`:                                     0,
		`<https://api/x?page=5>; rel=last`:                                       5,
		`malformed`:                                                              0,
	}
	for header, want := range cases {
		if got := lastLinkPage(header); got != want {
			t.Errorf("lastLinkPage(%q) = %d, want %d", header, got, want)
		}
	}
}
