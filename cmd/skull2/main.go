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

	reg := provider.NewDefaultRegistry()

	selected := make([]*config.Provider, 0, len(cfg.Providers))
	for i := range cfg.Providers {
		p := &cfg.Providers[i]
		if *providerName != "" && p.Name != *providerName {
			continue
		}
		selected = append(selected, p)
	}
	if *providerName != "" && len(selected) == 0 {
		return fmt.Errorf("unknown provider %q", *providerName)
	}

	ctx := context.Background()
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

func usage(w io.Writer) {
	fmt.Fprintf(w, `skull2 %s - multi-provider git portfolio manager

Usage:
  skull2 [command]

Commands:
  tui              Browse and clone repositories (default, planned)
  sync             Clone missing and fast-forward-pull existing repos (planned)
  refresh          Refresh the per-provider cache (planned)
  config           Show or validate configuration (planned)
  version          Print the version
  help             Show this help

See BRIEFING.md for the full design and 'beans list' for the build roadmap.
`, version)
}
