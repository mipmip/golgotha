// Command skull2 is a multi-provider git portfolio manager.
//
// This is the takeoff scaffolding entrypoint. Subcommands (config, refresh,
// sync, tui) are implemented across the OpenSpec changes tracked in beans.
package main

import (
	"fmt"
	"os"
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
	default:
		usage(os.Stderr)
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func usage(w *os.File) {
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
