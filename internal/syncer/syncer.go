// Package syncer clones missing repositories and fast-forward-pulls existing
// ones into the templated clone paths, skipping dirty trees.
//
// See BRIEFING.md section 10 for the sync semantics.
package syncer

import (
	"context"
	"fmt"

	"github.com/mipmip/huphop/internal/clonepath"
	"github.com/mipmip/huphop/internal/config"
	"github.com/mipmip/huphop/internal/provider"
)

// Action is the outcome recorded for a single repository.
type Action string

// Recorded actions per repository.
const (
	// ActionCloned means a missing repository was freshly cloned.
	ActionCloned Action = "cloned"
	// ActionUpdated means an existing clean clone was fast-forwarded.
	ActionUpdated Action = "updated"
	// ActionSkipped means an existing dirty clone was left untouched.
	ActionSkipped Action = "skipped"
	// ActionFailed means the repository could not be cloned or updated.
	ActionFailed Action = "failed"
)

// Result is the outcome for a single repository within a provider.
type Result struct {
	// Repo is the repository the action applied to.
	Repo provider.Repo
	// Path is the resolved local target path.
	Path string
	// Action is what the engine did.
	Action Action
	// Err is set when Action is ActionFailed.
	Err error
	// Warning carries a human-readable note (e.g. why a repo was skipped).
	Warning string
}

// ProviderSummary aggregates per-provider counts and per-repo results.
type ProviderSummary struct {
	// Provider is the provider name.
	Provider string
	// Cloned, Updated, Skipped and Failed are outcome counts.
	Cloned  int
	Updated int
	Skipped int
	Failed  int
	// Results holds the ordered per-repo results.
	Results []Result
}

// Summary is the full per-provider summary of a sync run.
type Summary struct {
	// Providers holds the per-provider summaries in run order.
	Providers []ProviderSummary
}

// Totals returns the summed counts across all providers.
func (s *Summary) Totals() (cloned, updated, skipped, failed int) {
	for _, p := range s.Providers {
		cloned += p.Cloned
		updated += p.Updated
		skipped += p.Skipped
		failed += p.Failed
	}
	return
}

// HasFailures reports whether any repository failed across all providers.
func (s *Summary) HasFailures() bool {
	_, _, _, failed := s.Totals()
	return failed > 0
}

// Engine reconciles the local filesystem with cached repositories.
type Engine struct {
	// Git is the git runner used for all operations.
	Git Git
	// Cfg is the loaded configuration (base dir, templates).
	Cfg *config.Config
}

// NewEngine builds an Engine with the given git runner and configuration.
func NewEngine(g Git, cfg *config.Config) *Engine {
	return &Engine{Git: g, Cfg: cfg}
}

// CloneRepo clones a single repository for provider p to its templated target
// path, reusing the same clone-URL, path resolution and git logic as a full
// sync. It is a no-op returning ActionUpdated=false semantics only insofar as it
// reports the result; callers (e.g. the TUI) use it to clone one repo on demand.
// If a repository already exists at the target it is left untouched and the
// returned Result carries ActionSkipped.
func (e *Engine) CloneRepo(ctx context.Context, p *config.Provider, r provider.Repo) Result {
	res := Result{Repo: r}

	target, err := clonepath.RenderFor(e.Cfg, p, p.WebURL, r.Owner, r.Name)
	if err != nil {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("resolving target path: %w", err)
		return res
	}
	res.Path = target

	if e.Git.IsRepo(target) {
		res.Action = ActionSkipped
		res.Warning = fmt.Sprintf("already cloned at %s", target)
		return res
	}

	url := cloneURL(p, r)
	if url == "" {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("no %s clone URL for %s/%s", p.CloneProtocol, r.Owner, r.Name)
		return res
	}
	if err := e.Git.Clone(ctx, url, target); err != nil {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("cloning %s: %w", url, err)
		return res
	}
	res.Action = ActionCloned
	return res
}

// SyncProvider reconciles one provider's repositories against the filesystem and
// returns its summary. Per-repo failures are collected and never abort the run.
func (e *Engine) SyncProvider(ctx context.Context, p *config.Provider, repos []provider.Repo) ProviderSummary {
	sum := ProviderSummary{Provider: p.Name}
	for i := range repos {
		res := e.syncRepo(ctx, p, repos[i])
		switch res.Action {
		case ActionCloned:
			sum.Cloned++
		case ActionUpdated:
			sum.Updated++
		case ActionSkipped:
			sum.Skipped++
		case ActionFailed:
			sum.Failed++
		}
		sum.Results = append(sum.Results, res)
	}
	return sum
}

// cloneURL returns the clone URL for repo under provider p based on the
// configured clone protocol.
func cloneURL(p *config.Provider, r provider.Repo) string {
	if p.CloneProtocol == config.ProtocolHTTPS {
		return r.HTTPSURL
	}
	return r.SSHURL
}

// syncRepo reconciles a single repository and returns its Result.
func (e *Engine) syncRepo(ctx context.Context, p *config.Provider, r provider.Repo) Result {
	res := Result{Repo: r}

	target, err := clonepath.RenderFor(e.Cfg, p, p.WebURL, r.Owner, r.Name)
	if err != nil {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("resolving target path: %w", err)
		return res
	}
	res.Path = target

	// Clone when there is no repository at the target yet.
	if !e.Git.IsRepo(target) {
		url := cloneURL(p, r)
		if url == "" {
			res.Action = ActionFailed
			res.Err = fmt.Errorf("no %s clone URL for %s/%s", p.CloneProtocol, r.Owner, r.Name)
			return res
		}
		if err := e.Git.Clone(ctx, url, target); err != nil {
			res.Action = ActionFailed
			res.Err = fmt.Errorf("cloning %s: %w", url, err)
			return res
		}
		res.Action = ActionCloned
		return res
	}

	// Existing repo: never touch dirty working trees.
	dirty, err := e.Git.IsDirty(ctx, target)
	if err != nil {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("checking status: %w", err)
		return res
	}
	if dirty {
		res.Action = ActionSkipped
		res.Warning = fmt.Sprintf("dirty working tree, skipping %s", target)
		return res
	}

	if err := e.Git.Fetch(ctx, target); err != nil {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("fetching: %w", err)
		return res
	}

	branch := r.DefaultBranch
	if branch == "" {
		branch, err = e.Git.CurrentDefaultBranch(ctx, target)
		if err != nil {
			res.Action = ActionFailed
			res.Err = fmt.Errorf("resolving default branch: %w", err)
			return res
		}
	}

	if err := e.Git.FastForward(ctx, target, branch); err != nil {
		res.Action = ActionFailed
		res.Err = fmt.Errorf("fast-forwarding %s: %w", branch, err)
		return res
	}
	res.Action = ActionUpdated
	return res
}
