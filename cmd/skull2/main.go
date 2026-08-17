// Command skull2 is a multi-provider git portfolio manager.
//
// This is the takeoff scaffolding entrypoint. Subcommands (config, refresh,
// sync, tui) are implemented across the OpenSpec changes tracked in beans.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mipmip/skull2/internal/config"
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
