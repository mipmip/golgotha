package tui

import (
	"context"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/skull2/internal/cache"
	"github.com/mipmip/skull2/internal/clonepath"
	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/provider"
	"github.com/mipmip/skull2/internal/syncer"
)

// cloneResultMsg is emitted after a single clone attempt completes.
type cloneResultMsg struct {
	Provider string
	Key      string
	Err      error
}

// bulkDoneMsg is emitted after a bulk clone finishes; it carries the per-repo
// results so the model can mark rows and report a summary.
type bulkDoneMsg struct {
	Results []cloneResultMsg
}

// openResultMsg is emitted after an open-in-browser attempt.
type openResultMsg struct {
	URL string
	Err error
}

// refreshResultMsg is emitted after a provider cache refresh completes.
type refreshResultMsg struct {
	Provider string
	Count    int
	Err      error
}

// ownerFetchedMsg is emitted after a lazy per-owner repo fetch completes.
type ownerFetchedMsg struct {
	Provider string
	Owner    string
	Repos    []provider.Repo
	Err      error
}

// cloneCmd clones one repo via the model's cloner and returns a cloneResultMsg.
func cloneCmd(c Cloner, p *config.Provider, r provider.Repo) tea.Cmd {
	return func() tea.Msg {
		res := c.CloneRepo(context.Background(), p, r)
		return cloneResultMsg{Provider: p.Name, Key: r.Owner + "/" + r.Name, Err: res.Err}
	}
}

// bulkCloneCmd clones several repos sequentially and returns a bulkDoneMsg.
func bulkCloneCmd(c Cloner, items []repoItem) tea.Cmd {
	return func() tea.Msg {
		var results []cloneResultMsg
		for _, it := range items {
			res := c.CloneRepo(context.Background(), it.Provider, it.Repo)
			results = append(results, cloneResultMsg{
				Provider: it.Provider.Name,
				Key:      it.key(),
				Err:      res.Err,
			})
		}
		return bulkDoneMsg{Results: results}
	}
}

// openCmd opens url in the browser via the stubbable openURL var.
func openCmd(url string) tea.Cmd {
	return func() tea.Msg {
		err := openURL(context.Background(), url)
		return openResultMsg{URL: url, Err: err}
	}
}

// defaultRefresher returns a refresher that re-fetches one provider via its
// client and rewrites the cache. It is never exercised in unit tests (no
// network); tests set m.refresher to nil or a stub.
func defaultRefresher(cfg *config.Config, m *Model) func(ctx context.Context, p *config.Provider) tea.Cmd {
	reg := provider.NewDefaultRegistry()
	return func(ctx context.Context, p *config.Provider) tea.Cmd {
		return func() tea.Msg {
			client, err := reg.Build(p)
			if err != nil {
				return refreshResultMsg{Provider: p.Name, Err: err}
			}

			// For all_owners providers the TUI stays lazy: re-discover the owner
			// index (cheap) and preserve per-owner fetch state; individual owners
			// are (re-)fetched on entry. For explicit-owners providers this does a
			// full repo fetch as before.
			if p.AllOwners {
				discovered, derr := client.ListOwners(ctx)
				if derr != nil {
					return refreshResultMsg{Provider: p.Name, Err: derr}
				}
				owners := config.ResolveOwners(p, discovered)
				c, _, lerr := cache.LoadOrEmpty(p.Name)
				if lerr != nil {
					return refreshResultMsg{Provider: p.Name, Err: lerr}
				}
				c.SetOwners(nowUTC(), owners)
				if serr := cache.Save(p.Name, c); serr != nil {
					return refreshResultMsg{Provider: p.Name, Err: serr}
				}
				return refreshResultMsg{Provider: p.Name, Count: len(c.Repos)}
			}

			repos, err := client.ListRepos(ctx, p.Owners)
			if err != nil {
				return refreshResultMsg{Provider: p.Name, Err: err}
			}
			if err := cache.Save(p.Name, cache.Cache{FetchedAt: nowUTC(), Repos: repos}); err != nil {
				return refreshResultMsg{Provider: p.Name, Err: err}
			}
			return refreshResultMsg{Provider: p.Name, Count: len(repos)}
		}
	}
}

// defaultOwnerFetcher returns a fetcher that lazily lists one owner's repos via
// its client and updates the provider cache's owner index. It is never exercised
// in unit tests (no network); tests set m.ownerFetcher to nil or a stub and feed
// ownerFetchedMsg directly.
func defaultOwnerFetcher(cfg *config.Config) func(ctx context.Context, p *config.Provider, owner string) tea.Cmd {
	reg := provider.NewDefaultRegistry()
	return func(ctx context.Context, p *config.Provider, owner string) tea.Cmd {
		return func() tea.Msg {
			client, err := reg.Build(p)
			if err != nil {
				return ownerFetchedMsg{Provider: p.Name, Owner: owner, Err: err}
			}
			var reqOwners []string
			if owner != config.SelfOwner {
				reqOwners = []string{owner}
			}
			repos, err := client.ListRepos(ctx, reqOwners)
			if err != nil {
				return ownerFetchedMsg{Provider: p.Name, Owner: owner, Err: err}
			}
			// Merge into the cache, marking this owner fetched.
			c, _, lerr := cache.LoadOrEmpty(p.Name)
			if lerr != nil {
				return ownerFetchedMsg{Provider: p.Name, Owner: owner, Err: lerr}
			}
			c.MarkOwnerFetched(owner, repos, nowUTC())
			if serr := cache.Save(p.Name, c); serr != nil {
				return ownerFetchedMsg{Provider: p.Name, Owner: owner, Err: serr}
			}
			return ownerFetchedMsg{Provider: p.Name, Owner: owner, Repos: repos}
		}
	}
}

// nowUTC returns the current UTC time (indirected for readability/testability).
func nowUTC() time.Time { return time.Now().UTC() }

// engineCloner adapts the syncer.Engine to the Cloner interface.
type engineCloner struct {
	eng *syncer.Engine
}

func newEngineCloner(cfg *config.Config) *engineCloner {
	return &engineCloner{eng: syncer.NewEngine(syncer.NewExecGit(), cfg)}
}

func (e *engineCloner) CloneRepo(ctx context.Context, p *config.Provider, r provider.Repo) syncerResult {
	res := e.eng.CloneRepo(ctx, p, r)
	return syncerResult{
		Target:  res.Path,
		Cloned:  res.Action == syncer.ActionCloned || res.Action == syncer.ActionSkipped,
		Err:     res.Err,
		Warning: res.Warning,
	}
}

// resolveTarget resolves the clone target path for a repo.
func resolveTarget(cfg *config.Config, p *config.Provider, r provider.Repo) (string, error) {
	return clonepath.RenderFor(cfg, p, p.WebURL, r.Owner, r.Name)
}

// isGitRepo reports whether target exists and contains a .git entry.
func isGitRepo(target string) bool {
	if target == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(target, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir() || info.Mode().IsRegular()
}
