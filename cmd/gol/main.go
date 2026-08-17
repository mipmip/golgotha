// Command gol is a multi-provider git portfolio manager.
//
// This is the takeoff scaffolding entrypoint. Subcommands (config, refresh,
// sync, tui) are implemented across the OpenSpec changes tracked in beans.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	golgotha "github.com/mipmip/golgotha"
	"github.com/mipmip/golgotha/internal/cache"
	"github.com/mipmip/golgotha/internal/config"
	"github.com/mipmip/golgotha/internal/provider"
	"github.com/mipmip/golgotha/internal/syncer"
	"github.com/mipmip/golgotha/internal/tui"
)

// version is set at build time via -ldflags "-X main.version=..." (goreleaser
// injects the git tag). When empty (dev builds), it falls back to the value
// embedded from the VERSION file — the single source of truth. See
// resolveVersion; the ldflags value wins when set.
var version = ""

// resolveVersion returns the ldflags-injected version when present, otherwise
// the version embedded from the VERSION file.
func resolveVersion() string {
	if version != "" {
		return version
	}
	return golgotha.Version
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gol:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}

	switch cmd {
	case "version", "--version", "-v":
		fmt.Println("gol", resolveVersion())
		return nil
	case "help", "-h", "--help":
		usage(os.Stdout)
		return nil
	case "", "tui":
		return runTUI()
	case "config":
		return runConfig(args[1:])
	case "refresh":
		return runRefresh(args[1:])
	case "sync":
		return runSync(args[1:])
	default:
		usage(os.Stderr)
		return fmt.Errorf("unknown command %q", cmd)
	}
}

// runTUI loads the configuration and launches the interactive Bubble Tea
// browser. It is the default command (bare `gol`) and `gol tui`.
func runTUI() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return tui.Run(cfg)
}

// runConfig handles the `config` subcommand: `path` and `check`.
func runConfig(args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}

	switch sub {
	case "path":
		path, err := config.DefaultPath()
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	case "check":
		path, err := config.DefaultPath()
		if err != nil {
			return err
		}
		cfg, err := config.LoadFrom(path)
		if err != nil {
			return err
		}
		fmt.Printf("config OK: %s\n", path)
		fmt.Printf("base_dir: %s\n", cfg.BaseDir)
		fmt.Printf("providers: %d\n", len(cfg.Providers))
		for i := range cfg.Providers {
			p := &cfg.Providers[i]
			fmt.Printf("  - %s (%s)\n", p.Name, p.Type)
		}
		return nil
	case "", "-h", "--help", "help":
		fmt.Println("Usage: gol config <path|check>")
		return nil
	default:
		return fmt.Errorf("unknown config subcommand %q", sub)
	}
}

// selectProviders returns the configured providers filtered to providerName when
// set. An unknown providerName yields an error.
func selectProviders(cfg *config.Config, providerName string) ([]*config.Provider, error) {
	selected := make([]*config.Provider, 0, len(cfg.Providers))
	for i := range cfg.Providers {
		p := &cfg.Providers[i]
		if providerName != "" && p.Name != providerName {
			continue
		}
		selected = append(selected, p)
	}
	if providerName != "" && len(selected) == 0 {
		return nil, fmt.Errorf("unknown provider %q", providerName)
	}
	return selected, nil
}

// refreshProviders re-fetches repositories for the selected providers and writes
// each per-provider JSON cache. It logs per-provider results and returns the
// first error encountered without aborting the run.
//
// When a provider has all_owners enabled it performs an eager discovery sweep:
// discover owners, resolve the effective owner set (self + discovered + explicit
// owners minus exclude_owners), and fetch repositories for every resolved owner,
// recording per-owner fetch state in the cache. Otherwise it preserves today's
// single ListRepos(p.Owners) behavior.
func refreshProviders(ctx context.Context, selected []*config.Provider) error {
	reg := provider.NewDefaultRegistry()
	printer := &cliProgressPrinter{w: os.Stdout}
	var firstErr error
	setErr := func(err error) {
		if firstErr == nil {
			firstErr = err
		}
	}
	for _, p := range selected {
		client, err := reg.Build(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gol: %s: %v\n", p.Name, err)
			setErr(err)
			continue
		}

		if p.AllOwners {
			if err := refreshAllOwners(ctx, client, p, printer); err != nil {
				fmt.Fprintf(os.Stderr, "gol: %s: %v\n", p.Name, err)
				setErr(err)
			}
			continue
		}

		if err := refreshExplicitOwners(ctx, client, p, printer); err != nil {
			fmt.Fprintf(os.Stderr, "gol: %s: %v\n", p.Name, err)
			setErr(err)
		}
	}
	return firstErr
}

// refreshExplicitOwners fetches the configured owners (or the authenticated
// user's own repos when none are configured) bounded-parallel with progress, and
// writes a flat provider cache. It returns the first fetch error encountered; a
// failed owner leaves the cache unwritten only when every owner failed.
func refreshExplicitOwners(ctx context.Context, client provider.Provider, p *config.Provider, printer *cliProgressPrinter) error {
	owners := p.Owners
	if len(owners) == 0 {
		owners = []string{config.SelfOwner}
	}

	results := fetchOwnersProgress(ctx, client, p, owners, printer)

	var (
		merged   []provider.Repo
		seen     = make(map[string]struct{})
		firstErr error
	)
	for _, r := range results {
		if r.Err != nil {
			if firstErr == nil {
				firstErr = r.Err
			}
			continue
		}
		for _, repo := range r.Repos {
			key := repo.Owner + "/" + repo.Name
			if _, dup := seen[key]; dup {
				continue
			}
			seen[key] = struct{}{}
			merged = append(merged, repo)
		}
	}

	if err := cache.Save(p.Name, cache.Cache{FetchedAt: time.Now().UTC(), Repos: merged}); err != nil {
		if firstErr == nil {
			firstErr = err
		}
	}
	fmt.Printf("%s: %d repos\n", p.Name, len(merged))
	return firstErr
}

// refreshAllOwners performs the eager discovery sweep for an all_owners provider:
// it discovers owners, resolves the effective set, fetches repos for every owner
// and writes a cache with a full owner index. Excluded owners are never fetched.
func refreshAllOwners(ctx context.Context, client provider.Provider, p *config.Provider, printer *cliProgressPrinter) error {
	discovered, err := client.ListOwners(ctx)
	if err != nil {
		return fmt.Errorf("discovering owners: %w", err)
	}
	if w := provider.ZeroDiscoveryWarning(p, discovered); w != "" {
		fmt.Fprintf(os.Stderr, "gol: warning: %s\n", w)
	}

	owners := config.ResolveOwners(p, discovered)
	now := time.Now().UTC()
	var c cache.Cache
	c.SetOwners(now, owners)

	results := fetchOwnersProgress(ctx, client, p, owners, printer)

	total := 0
	var firstErr error
	for _, r := range results {
		if r.Err != nil {
			// Commit-only-on-complete: a failed/canceled owner is left unfetched.
			if firstErr == nil {
				firstErr = fmt.Errorf("listing repos for owner %q: %w", displayOwner(r.Owner), r.Err)
			}
			continue
		}
		c.MarkOwnerFetched(r.Owner, r.Repos, time.Now().UTC())
		total += len(r.Repos)
	}

	if err := cache.Save(p.Name, c); err != nil {
		return err
	}
	fmt.Printf("%s: %d owners, %d repos\n", p.Name, len(owners), total)
	return firstErr
}

// displayOwner renders the SelfOwner sentinel for logs.
func displayOwner(owner string) string {
	if owner == config.SelfOwner {
		return "(self)"
	}
	return owner
}

// runRefresh handles the `refresh` subcommand: re-fetch repositories from each
// configured provider and update the per-provider JSON cache.
func runRefresh(args []string) error {
	fs := flag.NewFlagSet("refresh", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	providerName := fs.String("provider", "", "refresh only the named provider")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	selected, err := selectProviders(cfg, *providerName)
	if err != nil {
		return err
	}

	return refreshProviders(context.Background(), selected)
}

// runSync handles the `sync` subcommand: optionally refresh the cache, then
// clone missing and fast-forward-pull existing repositories for each selected
// provider. It logs one line per repository action plus a per-provider summary
// and exits non-zero if any repository failed.
func runSync(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	providerName := fs.String("provider", "", "sync only the named provider")
	noRefresh := fs.Bool("no-refresh", false, "skip refreshing the cache before syncing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	selected, err := selectProviders(cfg, *providerName)
	if err != nil {
		return err
	}

	ctx := context.Background()

	if !*noRefresh {
		// Refresh errors are logged but do not abort the sync; we still act on
		// whatever cache is available.
		if err := refreshProviders(ctx, selected); err != nil {
			fmt.Fprintf(os.Stderr, "gol: refresh reported errors: %v\n", err)
		}
	}

	engine := syncer.NewEngine(syncer.NewExecGit(), cfg)

	var summary syncer.Summary
	for _, p := range selected {
		c, ok, err := cache.LoadOrEmpty(p.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gol: %s: %v\n", p.Name, err)
			summary.Providers = append(summary.Providers, syncer.ProviderSummary{Provider: p.Name})
			continue
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "gol: %s: no cache; run 'gol refresh' first\n", p.Name)
			summary.Providers = append(summary.Providers, syncer.ProviderSummary{Provider: p.Name})
			continue
		}

		repos := provider.FilterRepos(p, c.Repos)
		ps := engine.SyncProvider(ctx, p, repos)
		logProviderSummary(ps)
		summary.Providers = append(summary.Providers, ps)
	}

	cloned, updated, skipped, failed := summary.Totals()
	fmt.Printf("total: %d cloned, %d updated, %d skipped, %d failed\n", cloned, updated, skipped, failed)

	if summary.HasFailures() {
		return fmt.Errorf("%d repositories failed", failed)
	}
	return nil
}

// logProviderSummary prints one line per repository action followed by a
// provider summary line.
func logProviderSummary(ps syncer.ProviderSummary) {
	for _, r := range ps.Results {
		name := r.Repo.Owner + "/" + r.Repo.Name
		switch r.Action {
		case syncer.ActionCloned:
			fmt.Printf("%s: cloned %s -> %s\n", ps.Provider, name, r.Path)
		case syncer.ActionUpdated:
			fmt.Printf("%s: updated %s\n", ps.Provider, name)
		case syncer.ActionSkipped:
			fmt.Printf("%s: skipped %s (%s)\n", ps.Provider, name, r.Warning)
			fmt.Fprintf(os.Stderr, "gol: %s: warning: %s\n", ps.Provider, r.Warning)
		case syncer.ActionFailed:
			fmt.Printf("%s: failed %s\n", ps.Provider, name)
			fmt.Fprintf(os.Stderr, "gol: %s: %s: %v\n", ps.Provider, name, r.Err)
		}
	}
	fmt.Printf("%s: %d cloned, %d updated, %d skipped, %d failed\n",
		ps.Provider, ps.Cloned, ps.Updated, ps.Skipped, ps.Failed)
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `gol %s - multi-provider git portfolio manager

Usage:
  gol [command]

Commands:
  tui              Browse and clone repositories (default)
  sync             Clone missing and fast-forward-pull existing repos
  refresh          Refresh the per-provider cache
  config           Show or validate configuration
  version          Print the version
  help             Show this help

See BRIEFING.md for the full design and 'beans list' for the build roadmap.
`, resolveVersion())
}
