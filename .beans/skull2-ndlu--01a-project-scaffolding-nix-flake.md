---
# skull2-ndlu
title: 01a Project scaffolding & Nix flake
status: completed
type: epic
priority: high
created_at: 2026-08-17T08:39:27Z
updated_at: 2026-08-17T08:44:15Z
parent: skull2-j0od
---

Go module + package layout, entrypoint, and a plain-nix flake (no flake-utils) covering supported architectures with dev shell, build, run and check.

## Tasks
- [x] `go.mod` (module path, Go 1.26+)
- [x] `cmd/skull2/main.go` entrypoint (prints version)
- [x] `internal/` package skeleton (config, template, provider, cache, sync, tui)
- [x] Plain `flake.nix`: `forAllSystems` over supported archs, no flake-utils
- [x] Flake outputs: `packages.default` (buildGoModule), `devShells.default`, `checks`
- [x] `nix build`, `nix run`, `nix develop`, `nix flake check` all work
- [x] `CLAUDE.md` with build/test/lint/loop conventions
- [x] `.gitignore` for Go + nix result

## Summary of Changes

Go module `github.com/mipmip/skull2` with `cmd/skull2` entrypoint and the internal package skeleton. Plain multi-arch `flake.nix` (no flake-utils) with `packages.default`, `devShells.default` and `checks` (build + go test). `nix build`, `nix run`, `nix develop` and `nix flake check` all pass. Added CLAUDE.md build guide, README and .gitignore updates.
