---
# skull2-ndlu
title: 01a Project scaffolding & Nix flake
status: in-progress
type: epic
priority: high
created_at: 2026-08-17T08:39:27Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-j0od
---

Go module + package layout, entrypoint, and a plain-nix flake (no flake-utils) covering supported architectures with dev shell, build, run and check.

## Tasks
- [ ] `go.mod` (module path, Go 1.26+)
- [ ] `cmd/skull2/main.go` entrypoint (prints version)
- [ ] `internal/` package skeleton (config, template, provider, cache, sync, tui)
- [ ] Plain `flake.nix`: `forAllSystems` over supported archs, no flake-utils
- [ ] Flake outputs: `packages.default` (buildGoModule), `devShells.default`, `checks`
- [ ] `nix build`, `nix run`, `nix develop`, `nix flake check` all work
- [ ] `CLAUDE.md` with build/test/lint/loop conventions
- [ ] `.gitignore` for Go + nix result
