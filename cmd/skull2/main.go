// Command skull2 is a multi-provider git portfolio manager.
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

	"github.com/mipmip/skull2/internal/cache"
	"github.com/mipmip/skull2/internal/config"
	"github.com/mipmip/skull2/internal/provider"
	"github.com/mipmip/skull2/internal/syncer"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "0.0.0-dev"

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "skull2:", err)
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
		fmt.Println("skull2", version)
		return nil
	case "", "help", "-h", "--help":
		usage(os.Stdout)
		return nil
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
		fmt.Println("Usage: skull2 config <path|check>")
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
func refreshProviders(ctx context.Context, selected []*config.Provider) error {
	reg := provider.NewDefaultRegistry()
	var firstErr error
	for _, p := range selected {
		client, err := reg.Build(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skull2: %s: %v\n", p.Name, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		repos, err := client.ListRepos(ctx, p.Owners)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skull2: %s: %v\n", p.Name, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := cache.Save(p.Name, cache.Cache{FetchedAt: time.Now().UTC(), Repos: repos}); err != nil {
			fmt.Fprintf(os.Stderr, "skull2: %s: %v\n", p.Name, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		fmt.Printf("%s: %d repos\n", p.Name, len(repos))
	}
	return firstErr
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
			fmt.Fprintf(os.Stderr, "skull2: refresh reported errors: %v\n", err)
		}
	}

	engine := syncer.NewEngine(syncer.NewExecGit(), cfg)

	var summary syncer.Summary
	for _, p := range selected {
		c, ok, err := cache.LoadOrEmpty(p.Name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "skull2: %s: %v\n", p.Name, err)
			summary.Providers = append(summary.Providers, syncer.ProviderSummary{Provider: p.Name})
			continue
		}
		if !ok {
			fmt.Fprintf(os.Stderr, "skull2: %s: no cache; run 'skull2 refresh' first\n", p.Name)
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
			fmt.Fprintf(os.Stderr, "skull2: %s: warning: %s\n", ps.Provider, r.Warning)
		case syncer.ActionFailed:
			fmt.Printf("%s: failed %s\n", ps.Provider, name)
			fmt.Fprintf(os.Stderr, "skull2: %s: %s: %v\n", ps.Provider, name, r.Err)
		}
	}
	fmt.Printf("%s: %d cloned, %d updated, %d skipped, %d failed\n",
		ps.Provider, ps.Cloned, ps.Updated, ps.Skipped, ps.Failed)
}

func usage(w io.Writer) {
	fmt.Fprintf(w, `skull2 %s - multi-provider git portfolio manager

Usage:
  skull2 [command]

Commands:
  tui              Browse and clone repositories (default, planned)
  sync             Clone missing and fast-forward-pull existing repos
  refresh          Refresh the per-provider cache (planned)
  config           Show or validate configuration (planned)
  version          Print the version
  help             Show this help

See BRIEFING.md for the full design and 'beans list' for the build roadmap.
`, version)
}
