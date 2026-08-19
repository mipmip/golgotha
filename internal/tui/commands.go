package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mipmip/huphop/internal/cache"
	"github.com/mipmip/huphop/internal/clonepath"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/fetch"
	"github.com/mipmip/huphop/internal/provider"
	"github.com/mipmip/huphop/internal/syncer"
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

// detailLoadedMsg is emitted after a repository's detail (tier-2 metadata +
// README) has been resolved — either loaded from the detail cache or freshly
// fetched. Cached reports whether it came from the cache (no network). When
// Err != nil and no cache existed, the model degrades gracefully (metadata plus
// a "README unavailable" note).
type detailLoadedMsg struct {
	Provider string
	Owner    string
	Name     string
	Details  cache.Details
	Cached   bool
	Err      error
}

// progressMsg carries one fetch progress event plus the channel it came from so
// the model can re-issue a wait for the next event, keeping Update pure.
type progressMsg struct {
	Provider string
	Owner    string
	Event    fetch.Event
	ch       <-chan fetch.Event
	// closed reports that the channel is drained (terminal); Event is zero.
	closed bool
}

// waitForProgress blocks on the next event from ch and wraps it in a
// progressMsg. When ch is closed it returns a progressMsg with closed=true so
// the model can finalize. It carries ch forward so Update can wait again.
func waitForProgress(provider, owner string, ch <-chan fetch.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return progressMsg{Provider: provider, Owner: owner, ch: ch, closed: true}
		}
		return progressMsg{Provider: provider, Owner: owner, Event: ev, ch: ch}
	}
}

// defaultProgressFetcher returns a streaming fetcher: it starts a page-aware
// fetch for one owner in a goroutine, forwarding progress events on the returned
// channel and closing it when done. The returned cancel stops the fetch. The
// cache is committed (MarkOwnerFetched) only on a complete, successful fetch —
// cancel or any failure leaves the owner unfetched. It is never exercised in
// unit tests (no network); tests inject a fake or feed progressMsg directly.
func defaultProgressFetcher(cfg *config.Config) func(ctx context.Context, p *config.Provider, owner string) (<-chan fetch.Event, context.CancelFunc) {
	reg := provider.NewDefaultRegistry()
	return progressFetcherWith(reg.Build)
}

// progressFetcherWith builds a streaming fetcher from an injectable provider
// constructor (the default uses the real registry; tests inject a fake).
//
// Ordering matters: FetchOwner emits its own terminal Done *before* it returns,
// but the owner is committed to the cache only after it returns. The UI handles
// Done by reloading from the cache, so a naive pass-through races the save and
// shows an empty list until restart. To avoid that we SUPPRESS the inner Done,
// commit the cache, then emit Done ourselves — guaranteeing the reload sees the
// saved repos. Failed/Canceled pass straight through (they don't touch the cache).
func progressFetcherWith(build func(*config.Provider) (provider.Provider, error)) func(ctx context.Context, p *config.Provider, owner string) (<-chan fetch.Event, context.CancelFunc) {
	return func(ctx context.Context, p *config.Provider, owner string) (<-chan fetch.Event, context.CancelFunc) {
		cctx, cancel := context.WithCancel(ctx)
		ch := make(chan fetch.Event, 16)

		go func() {
			defer close(ch)
			// raw delivers any event to the channel; emit is the same but drops
			// the inner Done (re-emitted after the cache commit).
			raw := fetch.Emit(func(ev fetch.Event) {
				select {
				case ch <- ev:
				case <-cctx.Done():
				}
			})
			emit := fetch.Emit(func(ev fetch.Event) {
				if ev.Kind == fetch.KindDone {
					return
				}
				raw(ev)
			})

			client, err := build(p)
			if err != nil {
				raw.Failed(p.Name, owner, err)
				return
			}
			of, ok := client.(provider.OwnerFetcher)
			if !ok {
				raw.Failed(p.Name, owner, fmt.Errorf("provider %q does not support progress fetch", p.Name))
				return
			}

			repos, ferr := of.FetchOwner(cctx, emit, owner)
			if ferr != nil {
				// FetchOwner already emitted Failed/Canceled; do not touch the cache.
				return
			}

			// Complete: commit the owner to the cache (all-or-nothing) BEFORE the
			// UI observes Done, so its reload-from-cache finds the repos.
			c, _, lerr := cache.LoadOrEmpty(p.Name)
			if lerr != nil {
				// Emit a terminal event so the UI leaves the fetching state.
				raw.Failed(p.Name, owner, fmt.Errorf("cache load failed: %w", lerr))
				return
			}
			c.MarkOwnerFetched(owner, repos, nowUTC())
			if serr := cache.Save(p.Name, c); serr != nil {
				raw.Warning(p.Name, owner, "cache save failed: "+serr.Error())
			}
			raw.Done(p.Name, owner, len(repos))
		}()

		return ch, cancel
	}
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
			repos, err := client.ListRepos(ctx, []string{owner})
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

// defaultDetailFetcher returns a fetcher that resolves a repository's detail
// (tier-2 metadata + README). It fetches from the provider and writes the
// separate per-repo detail cache, returning a detailLoadedMsg. On a fetch error
// it falls back to any existing detail cache; if none exists the message carries
// the error so the model degrades gracefully. It is never exercised in unit
// tests (no network); tests set m.detailFetcher to nil or feed detailLoadedMsg.
func defaultDetailFetcher(cfg *config.Config) func(ctx context.Context, p *config.Provider, r provider.Repo) tea.Cmd {
	reg := provider.NewDefaultRegistry()
	return func(ctx context.Context, p *config.Provider, r provider.Repo) tea.Cmd {
		return func() tea.Msg {
			msg := detailLoadedMsg{Provider: p.Name, Owner: r.Owner, Name: r.Name}

			client, err := reg.Build(p)
			if err != nil {
				return detailFallback(msg, err)
			}
			pd, derr := client.RepoDetails(ctx, r.Owner, r.Name)
			if derr != nil {
				return detailFallback(msg, derr)
			}
			readme, rerr := client.Readme(ctx, r.Owner, r.Name)
			if rerr != nil {
				return detailFallback(msg, rerr)
			}

			d, serr := cache.RefreshDetails(p.Name, r.Owner, r.Name, pd, readme, nowUTC())
			if serr != nil {
				// Details fetched fine but caching failed; still show them.
				msg.Details = cache.Details{
					FetchedAt:      nowUTC(),
					Stars:          pd.Stars,
					Topics:         pd.Topics,
					Language:       pd.Language,
					ReadmeMarkdown: readme,
				}
				return msg
			}
			msg.Details = d
			return msg
		}
	}
}

// detailFallback turns a fetch error into a detailLoadedMsg: it uses an existing
// detail cache when present (Cached=true, no error), otherwise carries the error
// so the model shows the graceful-offline note.
func detailFallback(msg detailLoadedMsg, fetchErr error) tea.Msg {
	if d, ok, lerr := cache.LoadDetailsOrEmpty(msg.Provider, msg.Owner, msg.Name); lerr == nil && ok {
		msg.Details = d
		msg.Cached = true
		return msg
	}
	msg.Err = fetchErr
	return msg
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

// cloneEvent is a streaming clone progress/terminal event for the multiplex
// clone popup.
type cloneEvent struct {
	Frac  float64
	Phase string
	Done  bool
	Err   error
}

// cloneMsg carries one cloneEvent plus the channel to keep listening on.
type cloneMsg struct {
	ev cloneEvent
	ch <-chan cloneEvent
}

// waitForClone blocks on the next clone event and wraps it in a cloneMsg.
func waitForClone(ch <-chan cloneEvent) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return cloneMsg{ev: cloneEvent{Done: true}, ch: ch}
		}
		return cloneMsg{ev: ev, ch: ch}
	}
}

// defaultProgressCloner returns a streaming clone seam backed by the syncer
// engine (which emits git --progress fractions).
func defaultProgressCloner(cfg *config.Config) func(ctx context.Context, p *config.Provider, r provider.Repo) (<-chan cloneEvent, context.CancelFunc) {
	eng := syncer.NewEngine(syncer.NewExecGit(), cfg)
	return func(ctx context.Context, p *config.Provider, r provider.Repo) (<-chan cloneEvent, context.CancelFunc) {
		ctx, cancel := context.WithCancel(ctx)
		ch := make(chan cloneEvent, 16)
		go func() {
			defer close(ch)
			res := eng.CloneRepoProgress(ctx, p, r, func(frac float64, phase string) {
				select {
				case ch <- cloneEvent{Frac: frac, Phase: phase}:
				case <-ctx.Done():
				}
			})
			if res.Err != nil {
				ch <- cloneEvent{Done: true, Err: res.Err}
				return
			}
			ch <- cloneEvent{Done: true}
		}()
		return ch, cancel
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
