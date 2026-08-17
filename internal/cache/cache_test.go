package cache

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/mipmip/skull2/internal/provider"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	in := Cache{
		FetchedAt: time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC),
		Repos: []provider.Repo{
			{
				Owner: "acme", Name: "alpha", Description: "d",
				SSHURL: "git@x:acme/alpha.git", HTTPSURL: "https://x/acme/alpha.git",
				WebURL: "https://x/acme/alpha", DefaultBranch: "main",
				Archived: false, Fork: true,
				UpdatedAt: time.Date(2023, 6, 7, 8, 9, 10, 0, time.UTC),
			},
			{Owner: "acme", Name: "beta"},
		},
	}

	if err := Save("github", in); err != nil {
		t.Fatal(err)
	}
	got, err := Load("github")
	if err != nil {
		t.Fatal(err)
	}
	if !got.FetchedAt.Equal(in.FetchedAt) {
		t.Fatalf("fetched_at: got %v want %v", got.FetchedAt, in.FetchedAt)
	}
	if !reflect.DeepEqual(got.Repos, in.Repos) {
		t.Fatalf("repos mismatch:\n got %+v\nwant %+v", got.Repos, in.Repos)
	}

	// File lives at <dir>/skull2/github.json.
	if _, err := os.Stat(filepath.Join(dir, "skull2", "github.json")); err != nil {
		t.Fatalf("cache file not at expected path: %v", err)
	}
}

func TestSaveAtomicNoTempLeftBehind(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	if err := Save("gitlab", Cache{FetchedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "skull2"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
	if len(entries) != 1 || entries[0].Name() != "gitlab.json" {
		t.Fatalf("unexpected cache dir contents: %v", entries)
	}
}

func TestLoadMissingIsError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	if _, err := Load("nope"); err == nil {
		t.Fatal("expected error for missing cache")
	}
}

func TestLoadOrEmptyMissingTolerated(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	c, ok, err := LoadOrEmpty("nope")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for missing cache")
	}
	if len(c.Repos) != 0 || !c.FetchedAt.IsZero() {
		t.Fatalf("expected empty cache, got %+v", c)
	}
}

func TestLoadOrEmptyExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	if err := Save("cb", Cache{FetchedAt: time.Unix(1000, 0).UTC(), Repos: []provider.Repo{{Name: "r"}}}); err != nil {
		t.Fatal(err)
	}
	c, ok, err := LoadOrEmpty("cb")
	if err != nil {
		t.Fatal(err)
	}
	if !ok || len(c.Repos) != 1 || c.Repos[0].Name != "r" {
		t.Fatalf("existing cache not loaded: ok=%v c=%+v", ok, c)
	}
}

func TestLoadCorruptJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	cdir := filepath.Join(dir, "skull2")
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cdir, "bad.json"), []byte("{not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load("bad"); err == nil {
		t.Fatal("expected parse error")
	}
	// LoadOrEmpty surfaces non-not-exist errors too.
	if _, _, err := LoadOrEmpty("bad"); err == nil {
		t.Fatal("expected parse error from LoadOrEmpty")
	}
}

func TestSaveMkdirError(t *testing.T) {
	dir := t.TempDir()
	// Make the XDG cache root a regular file so MkdirAll(<file>/skull2) fails.
	f := filepath.Join(dir, "asfile")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CACHE_HOME", f)
	if err := Save("github", Cache{FetchedAt: time.Now()}); err == nil {
		t.Fatal("expected mkdir error")
	}
}

func TestSaveThenLoadDefaultProviderMissingReadError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	// Create the provider path as a directory so ReadFile fails with a non
	// not-exist error, exercising Load's read-error branch.
	cdir := filepath.Join(dir, "skull2")
	if err := os.MkdirAll(filepath.Join(cdir, "dirprov.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := Load("dirprov"); err == nil {
		t.Fatal("expected read error for directory path")
	}
	if _, _, err := LoadOrEmpty("dirprov"); err == nil {
		t.Fatal("expected read error from LoadOrEmpty")
	}
}

func TestSaveRenameError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	cdir := filepath.Join(dir, "skull2")
	// Pre-create the final path as a directory so os.Rename(tmpfile, final) fails.
	if err := os.MkdirAll(filepath.Join(cdir, "github.json"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Save("github", Cache{FetchedAt: time.Now()}); err == nil {
		t.Fatal("expected rename error when final path is a directory")
	}
	// No temp file must be left behind after the failed rename.
	entries, err := os.ReadDir(cdir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("temp file left behind after failed rename: %s", e.Name())
		}
	}
}

func TestSaveOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	if err := Save("gh", Cache{Repos: []provider.Repo{{Name: "first"}}}); err != nil {
		t.Fatal(err)
	}
	if err := Save("gh", Cache{Repos: []provider.Repo{{Name: "second"}}}); err != nil {
		t.Fatal(err)
	}
	got, err := Load("gh")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Repos) != 1 || got.Repos[0].Name != "second" {
		t.Fatalf("save did not overwrite: %+v", got.Repos)
	}
}

func TestSaveTempCreateError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	cdir := filepath.Join(dir, "skull2")
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Remove write/execute perms so CreateTemp inside the dir fails.
	if err := os.Chmod(cdir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(cdir, 0o755) })
	err := Save("gh", Cache{FetchedAt: time.Now()})
	if err == nil {
		// Running as root ignores mode bits; skip rather than fail spuriously.
		t.Skip("temp-file creation succeeded despite mode; likely running as root")
	}
}

func TestDirErrorWhenHomeUnresolvable(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "")
	// On unix, an empty HOME makes os.UserHomeDir return an error.
	if _, err := Dir(); err == nil {
		t.Skip("home resolvable in this environment; skipping error-path check")
	}
	if _, err := Path("x"); err == nil {
		t.Fatal("expected Path to propagate Dir error")
	}
}

func TestDirFallbackToHome(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	home := t.TempDir()
	t.Setenv("HOME", home)
	d, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if d != filepath.Join(home, ".cache", "skull2") {
		t.Fatalf("dir = %q", d)
	}
}

func TestPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	p, err := Path("github")
	if err != nil {
		t.Fatal(err)
	}
	if p != filepath.Join(dir, "skull2", "github.json") {
		t.Fatalf("path = %q", p)
	}
}

func TestOwnerIndexRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	now := time.Date(2024, 5, 6, 7, 8, 9, 0, time.UTC)
	var c Cache
	c.SetOwners(now, []string{"", "acme", "beta"}) // "" = SelfOwner
	c.MarkOwnerFetched("acme", []provider.Repo{{Owner: "acme", Name: "alpha"}}, now)

	if err := Save("gh", c); err != nil {
		t.Fatal(err)
	}
	got, err := Load("gh")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Owners) != 3 {
		t.Fatalf("owners = %+v", got.Owners)
	}
	if !got.OwnerFetched("acme") {
		t.Fatal("acme should be fetched")
	}
	if got.OwnerFetched("beta") {
		t.Fatal("beta should be unfetched")
	}
	if unf := got.UnfetchedOwners(); len(unf) != 2 {
		t.Fatalf("unfetched = %v, want self+beta", unf)
	}
	if repos := got.ReposFor("acme"); len(repos) != 1 || repos[0].Name != "alpha" {
		t.Fatalf("acme repos = %+v", repos)
	}
}

func TestMarkOwnerFetchedUpdatesOnlyThatOwner(t *testing.T) {
	now := time.Now().UTC()
	var c Cache
	c.SetOwners(now, []string{"a", "b"})
	c.MarkOwnerFetched("a", []provider.Repo{{Owner: "a", Name: "a1"}}, now)
	c.MarkOwnerFetched("b", []provider.Repo{{Owner: "b", Name: "b1"}}, now)

	// Re-fetch a with new repos; b must be untouched.
	c.MarkOwnerFetched("a", []provider.Repo{{Owner: "a", Name: "a2"}}, now)

	if repos := c.ReposFor("a"); len(repos) != 1 || repos[0].Name != "a2" {
		t.Fatalf("a repos = %+v", repos)
	}
	if repos := c.ReposFor("b"); len(repos) != 1 || repos[0].Name != "b1" {
		t.Fatalf("b repos = %+v", repos)
	}
}

// TestCommitOnlyOnCompleteLeavesFailedOwnerUnfetched models the repo-cache
// contract: an owner whose fetch failed/canceled (so MarkOwnerFetched is not
// called for it) stays unfetched with no repos, while a completed owner commits.
func TestCommitOnlyOnCompleteLeavesFailedOwnerUnfetched(t *testing.T) {
	now := time.Now().UTC()
	var c Cache
	c.SetOwners(now, []string{"good", "bad"})

	// Only the completed owner is committed; the failed owner is left as-is.
	c.MarkOwnerFetched("good", []provider.Repo{{Owner: "good", Name: "g1"}}, now)

	if !c.OwnerFetched("good") {
		t.Fatal("completed owner must be fetched")
	}
	if c.OwnerFetched("bad") {
		t.Fatal("uncommitted owner must remain unfetched")
	}
	if len(c.ReposFor("bad")) != 0 {
		t.Fatal("uncommitted owner must have no repos")
	}
	if got := c.UnfetchedOwners(); len(got) != 1 || got[0] != "bad" {
		t.Fatalf("UnfetchedOwners = %v, want [bad]", got)
	}
}

func TestSetOwnersDropsRemovedOwnerRepos(t *testing.T) {
	now := time.Now().UTC()
	var c Cache
	c.SetOwners(now, []string{"a", "b"})
	c.MarkOwnerFetched("a", []provider.Repo{{Owner: "a", Name: "a1"}}, now)
	c.MarkOwnerFetched("b", []provider.Repo{{Owner: "b", Name: "b1"}}, now)

	// Rediscover without b: its repos and state are dropped; a's fetch preserved.
	c.SetOwners(now, []string{"a"})
	if len(c.Owners) != 1 || c.Owners[0].Name != "a" || c.Owners[0].FetchedAt == nil {
		t.Fatalf("owners after rediscovery = %+v", c.Owners)
	}
	if len(c.ReposFor("b")) != 0 {
		t.Fatal("b repos should be dropped")
	}
	if len(c.ReposFor("a")) != 1 {
		t.Fatal("a repos should be preserved")
	}
}

func TestLegacyFlatCacheRead(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	cdir := filepath.Join(dir, "skull2")
	if err := os.MkdirAll(cdir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Legacy flat file: no owner index.
	legacy := `{"fetched_at":"2024-01-02T03:04:05Z","repos":[
		{"Owner":"acme","Name":"alpha"},
		{"Owner":"acme","Name":"beta"},
		{"Owner":"pim","Name":"dots"}
	]}`
	if err := os.WriteFile(filepath.Join(cdir, "legacy.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load("legacy")
	if err != nil {
		t.Fatal(err)
	}
	// Every owner present in repos is treated as fetched.
	if len(c.Owners) != 2 {
		t.Fatalf("synthesized owners = %+v", c.Owners)
	}
	if !c.OwnerFetched("acme") || !c.OwnerFetched("pim") {
		t.Fatal("legacy owners should be treated as fetched")
	}
	if len(c.UnfetchedOwners()) != 0 {
		t.Fatalf("legacy cache should have no unfetched owners: %v", c.UnfetchedOwners())
	}
	if len(c.ReposFor("acme")) != 2 {
		t.Fatalf("acme repos = %+v", c.ReposFor("acme"))
	}
}
