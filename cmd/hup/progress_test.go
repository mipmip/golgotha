package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
	"github.com/mipmip/huphop/internal/provider"
)

// fakeFetchClient implements provider.Provider + provider.OwnerFetcher. It emits
// a Started/PageDone/Done (or Failed) sequence per owner and can fail a named
// owner. It records max concurrency to assert bounded parallelism.
type fakeFetchClient struct {
	owners  []string
	failOn  string
	mu      sync.Mutex
	inFlt   int32
	maxFlt  int32
	fetched []string
}

func (f *fakeFetchClient) ListOwners(_ context.Context) ([]string, error) { return f.owners, nil }

func (f *fakeFetchClient) RepoDetails(_ context.Context, _, _ string) (provider.Details, error) {
	return provider.Details{}, nil
}

func (f *fakeFetchClient) Readme(_ context.Context, _, _ string) (string, error) { return "", nil }

func (f *fakeFetchClient) ListRepos(ctx context.Context, owners []string) ([]provider.Repo, error) {
	owner := ""
	if len(owners) > 0 {
		owner = owners[0]
	}
	return f.FetchOwner(ctx, nil, owner)
}

func (f *fakeFetchClient) FetchOwner(ctx context.Context, emit fetch.Emit, owner string) ([]provider.Repo, error) {
	cur := atomic.AddInt32(&f.inFlt, 1)
	for {
		old := atomic.LoadInt32(&f.maxFlt)
		if cur <= old || atomic.CompareAndSwapInt32(&f.maxFlt, old, cur) {
			break
		}
	}
	defer atomic.AddInt32(&f.inFlt, -1)

	emit.Started("prov", owner)
	if owner == f.failOn {
		err := errors.New("simulated failure")
		emit.Failed("prov", owner, err)
		return nil, err
	}
	emit.PageDone("prov", owner, 1, 1, 1)

	f.mu.Lock()
	f.fetched = append(f.fetched, owner)
	f.mu.Unlock()

	repos := []provider.Repo{{Owner: ownerOrSelf(owner), Name: ownerOrSelf(owner) + "-repo"}}
	emit.Done("prov", owner, len(repos))
	return repos, nil
}

func ownerOrSelf(o string) string {
	if o == "" {
		return "me"
	}
	return o
}

func TestFetchOwnersProgressPrintsPerOwnerLines(t *testing.T) {
	p := &config.Provider{Name: "prov", Type: config.ProviderGitHub}
	client := &fakeFetchClient{}
	var buf bytes.Buffer
	printer := &cliProgressPrinter{w: &buf}

	results := fetchOwnersProgress(context.Background(), client, p, []string{"acme", "beta"}, printer)
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	out := buf.String()
	for _, want := range []string{
		"prov: acme: fetching...",
		"prov: acme: fetched 1 repos",
		"prov: beta: fetching...",
		"prov: beta: fetched 1 repos",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q\n---\n%s", want, out)
		}
	}
}

func TestFetchOwnersProgressBoundedConcurrency(t *testing.T) {
	p := &config.Provider{Name: "prov", Type: config.ProviderGitHub}
	client := &fakeFetchClient{}
	printer := &cliProgressPrinter{w: &bytes.Buffer{}}

	owners := make([]string, 20)
	for i := range owners {
		owners[i] = fmt.Sprintf("o%d", i)
	}
	fetchOwnersProgress(context.Background(), client, p, owners, printer)
	if client.maxFlt > int32(fetch.WorkerCap) {
		t.Fatalf("max in-flight owners = %d, want <= %d", client.maxFlt, fetch.WorkerCap)
	}
	if client.maxFlt < 1 {
		t.Fatal("no owners fetched")
	}
}

func TestRefreshAllOwnersCommitOnlyOnComplete(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	p := &config.Provider{
		Name:          "prov",
		Type:          config.ProviderGitHub,
		Username:      "me",
		AllOwners:     true,
		ExcludeOwners: []string{"me"}, // exclude self; keep it to discovered orgs only
	}
	client := &fakeFetchClient{owners: []string{"good", "bad"}, failOn: "bad"}
	printer := &cliProgressPrinter{w: &bytes.Buffer{}}

	err := refreshAllOwners(context.Background(), client, p, printer)
	if err == nil {
		t.Fatal("expected an error from the failing owner")
	}

	c, lerr := cache.Load("prov")
	if lerr != nil {
		t.Fatal(lerr)
	}
	// good committed (fetched), bad left unfetched (in index but FetchedAt nil).
	if !c.OwnerFetched("good") {
		t.Fatal("good should be committed as fetched")
	}
	if c.OwnerFetched("bad") {
		t.Fatal("bad failed; it must NOT be committed as fetched")
	}
	if len(c.ReposFor("bad")) != 0 {
		t.Fatal("failed owner must have no repos in the cache")
	}
	if len(c.ReposFor("good")) == 0 {
		t.Fatal("good owner should have repos in the cache")
	}
}

func TestRefreshExplicitOwnersWritesCache(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	p := &config.Provider{Name: "prov", Type: config.ProviderGitHub, Owners: []string{"acme", "beta"}}
	client := &fakeFetchClient{}
	printer := &cliProgressPrinter{w: &bytes.Buffer{}}

	if err := refreshExplicitOwners(context.Background(), client, p, printer); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Load("prov")
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Repos) != 2 {
		t.Fatalf("got %d repos, want 2: %+v", len(c.Repos), c.Repos)
	}
}
