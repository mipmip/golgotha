package cache

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/mipmip/skull2/internal/provider"
)

func TestDetailsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	want := Details{
		FetchedAt:      time.Now().UTC().Truncate(time.Second),
		Stars:          123,
		Topics:         []string{"cli", "go"},
		Language:       "Go",
		ReadmeMarkdown: "# Title\n\nbody\n",
	}
	if err := SaveDetails("github", "acme", "alpha", want); err != nil {
		t.Fatal(err)
	}

	got, ok, err := LoadDetailsOrEmpty("github", "acme", "alpha")
	if err != nil || !ok {
		t.Fatalf("load: ok=%v err=%v", ok, err)
	}
	if got.Stars != want.Stars || got.Language != want.Language ||
		got.ReadmeMarkdown != want.ReadmeMarkdown || len(got.Topics) != 2 ||
		!got.FetchedAt.Equal(want.FetchedAt) {
		t.Fatalf("round-trip mismatch: got %+v want %+v", got, want)
	}
}

func TestDetailsMissing(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	got, ok, err := LoadDetailsOrEmpty("github", "acme", "nope")
	if err != nil {
		t.Fatalf("missing should not error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for missing detail cache")
	}
	if got.Stars != 0 || got.ReadmeMarkdown != "" {
		t.Fatalf("expected zero details, got %+v", got)
	}
}

func TestRefreshDetailsWritesAndReturns(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	pd := provider.Details{Stars: 9, Topics: []string{"x"}, Language: "Rust"}
	ts := time.Now().UTC().Truncate(time.Second)

	d, err := RefreshDetails("gitlab", "grp/sub", "proj", pd, "raw md", ts)
	if err != nil {
		t.Fatal(err)
	}
	if d.Stars != 9 || d.Language != "Rust" || d.ReadmeMarkdown != "raw md" || !d.FetchedAt.Equal(ts) {
		t.Fatalf("returned details wrong: %+v", d)
	}

	// It must have persisted and be loadable (nested group flattened to one file).
	got, ok, err := LoadDetailsOrEmpty("gitlab", "grp/sub", "proj")
	if err != nil || !ok {
		t.Fatalf("refresh did not persist: ok=%v err=%v", ok, err)
	}
	if got.Stars != 9 || got.ReadmeMarkdown != "raw md" {
		t.Fatalf("persisted details wrong: %+v", got)
	}

	// A second refresh overwrites (round-trips the new value).
	pd2 := provider.Details{Stars: 20, Language: "Go"}
	if _, err := RefreshDetails("gitlab", "grp/sub", "proj", pd2, "new", ts.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	got2, _, _ := LoadDetailsOrEmpty("gitlab", "grp/sub", "proj")
	if got2.Stars != 20 || got2.ReadmeMarkdown != "new" {
		t.Fatalf("refresh did not overwrite: %+v", got2)
	}
}

func TestDetailsSeparateFromListCache(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())

	// Write a list cache, then a detail cache; the list cache must be untouched.
	list := Cache{FetchedAt: time.Now().UTC(), Repos: []provider.Repo{{Owner: "acme", Name: "alpha"}}}
	if err := Save("github", list); err != nil {
		t.Fatal(err)
	}
	if err := SaveDetails("github", "acme", "alpha", Details{Stars: 1}); err != nil {
		t.Fatal(err)
	}

	// The detail file lives under details/<provider>/, not the provider file.
	dp, _ := DetailsPath("github", "acme", "alpha")
	if filepath.Base(filepath.Dir(dp)) != "github" ||
		filepath.Base(filepath.Dir(filepath.Dir(dp))) != "details" {
		t.Fatalf("detail path not under details/github: %s", dp)
	}

	got, ok, err := LoadOrEmpty("github")
	if err != nil || !ok {
		t.Fatalf("list cache load: ok=%v err=%v", ok, err)
	}
	if len(got.Repos) != 1 || got.Repos[0].Name != "alpha" {
		t.Fatalf("list cache modified by detail write: %+v", got.Repos)
	}
}

func TestDetailFileNameSanitizes(t *testing.T) {
	// Nested owner (GitLab group) must stay a single flat file segment.
	name := detailFileName("group/sub", "proj")
	if filepath.Base(name) != name {
		t.Fatalf("detail file name not flat: %q", name)
	}
	if name != "group-sub__proj.json" {
		t.Fatalf("unexpected sanitized name: %q", name)
	}
}
