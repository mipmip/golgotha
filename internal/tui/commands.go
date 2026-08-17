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
