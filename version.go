// Package huphop exposes the embedded release version as the single source of
// truth. The VERSION file at the repo root is read at build time via go:embed;
// the same file is read by flake.nix (builtins.readFile) and by goreleaser
// (git tag → ldflags override). Keeping one file with three consumers prevents
// version drift.
package huphop

import (
	_ "embed"
	"strings"
)

//go:embed VERSION
var rawVersion string

// Version is the trimmed release version from the VERSION file, e.g. "0.1.0".
var Version = strings.TrimSpace(rawVersion)
