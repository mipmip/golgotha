package provider

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/fetch"
)

// OwnerFetcher is a provider client that can fetch a single owner's
// repositories while emitting progress events. The built-in GitHub, Codeberg
// and GitLab clients all satisfy it; callers that want progress (the TUI and the
// CLI) type-assert a Provider to OwnerFetcher.
type OwnerFetcher interface {
	// FetchOwner fetches one owner's repositories page by page, emitting progress
	// through emit (which may be nil). An empty owner (config.SelfOwner) fetches
	// the authenticated user's own repositories. On cancellation it returns
	// ctx.Err() after emitting a Canceled event; on failure it returns the error
	// after a Failed event.
	FetchOwner(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error)
}

// repoKey is the owner/name dedupe key used when merging pages and owners.
func repoKey(r Repo) string { return r.Owner + "/" + r.Name }

// listReposOverFetch implements the plain, event-free ListRepos contract over an
// owner-at-a-time FetchOwner. An empty owners slice means the authenticated
// user's own repositories (config.SelfOwner). Results across owners are merged
// and deduped by owner/name; per-owner filtering already happened in FetchOwner,
// so the merged result is returned as-is.
func listReposOverFetch(
	ctx context.Context,
	owners []string,
	_ *config.Provider,
	fetchOwner func(ctx context.Context, emit fetch.Emit, owner string) ([]Repo, error),
) ([]Repo, error) {
	req := owners
	if len(req) == 0 {
		req = []string{config.SelfOwner}
	}

	var all []Repo
	seen := make(map[string]struct{})
	for _, owner := range req {
		repos, err := fetchOwner(ctx, nil, owner)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			key := repoKey(r)
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, r)
		}
	}
	return all, nil
}

// osLookupEnv is the production EnvLookup; it mirrors os.LookupEnv.
func osLookupEnv(key string) (string, bool) { return os.LookupEnv(key) }

// readAllAndClose reads the full response body and closes it, always closing the
// body even when reading fails.
func readAllAndClose(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// NewDefaultRegistry returns a Registry with the built-in GitHub, Codeberg and
// GitLab clients registered. Constructors use http.DefaultClient and resolve
// auth via the real CLIs and environment.
func NewDefaultRegistry() *Registry {
	reg := NewRegistry()
	reg.Register(config.ProviderGitHub, func(p *config.Provider) (Provider, error) {
		return NewGitHub(*p, nil), nil
	})
	reg.Register(config.ProviderCodeberg, func(p *config.Provider) (Provider, error) {
		return NewCodeberg(*p, nil), nil
	})
	reg.Register(config.ProviderGitLab, func(p *config.Provider) (Provider, error) {
		return NewGitLab(*p, nil), nil
	})
	return reg
}
