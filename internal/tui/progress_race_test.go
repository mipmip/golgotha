package tui

import (
	"context"
	"testing"

	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
	"github.com/mipmip/huphop/internal/provider"
)

// fakeOwnerFetcher implements provider.Provider (+ OwnerFetcher). Its FetchOwner
// emits its own terminal Done (like the real page fetcher does) so the test can
// verify progressFetcherWith suppresses it and re-emits Done after committing.
type fakeOwnerFetcher struct{ repos []provider.Repo }

func (f *fakeOwnerFetcher) ListRepos(context.Context, []string) ([]provider.Repo, error) {
	return f.repos, nil
}
func (f *fakeOwnerFetcher) ListOwners(context.Context) ([]string, error) { return nil, nil }
func (f *fakeOwnerFetcher) RepoDetails(context.Context, string, string) (provider.Details, error) {
	return provider.Details{}, nil
}
func (f *fakeOwnerFetcher) Readme(context.Context, string, string) (string, error) { return "", nil }
func (f *fakeOwnerFetcher) FetchOwner(_ context.Context, emit fetch.Emit, owner string) ([]provider.Repo, error) {
	emit.Started("test", owner)
	emit.Done("test", owner, len(f.repos)) // inner Done — must be suppressed by the wrapper
	return f.repos, nil
}

// TestProgressFetcherCommitsBeforeDone asserts the fix: when the UI observes the
// Done event, the cache already contains the fetched owner's repos (so the
// reload-from-cache on Done shows them without a restart), and exactly one Done
// is delivered.
func TestProgressFetcherCommitsBeforeDone(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	p := &config.Provider{Name: "test", Type: "github"}
	repos := []provider.Repo{{Owner: "acme", Name: "r1"}, {Owner: "acme", Name: "r2"}}
	build := func(*config.Provider) (provider.Provider, error) {
		return &fakeOwnerFetcher{repos: repos}, nil
	}

	ch, cancel := progressFetcherWith(build)(context.Background(), p, "acme")
	defer cancel()

	dones := 0
	for ev := range ch {
		if ev.Kind == fetch.KindDone {
			dones++
			c, _, err := cache.LoadOrEmpty(p.Name)
			if err != nil {
				t.Fatalf("cache load: %v", err)
			}
			if got := len(c.ReposFor("acme")); got != 2 {
				t.Fatalf("cache not committed before Done: ReposFor(acme)=%d, want 2", got)
			}
		}
	}
	if dones != 1 {
		t.Fatalf("want exactly one Done event, got %d", dones)
	}
}
