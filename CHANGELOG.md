# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Required per-provider `username` field identifying your own account on each
  provider.
- TUI modes: `default_mode` + `modes:` configure per-mode chrome as ordered
  `header`/`footer` element slots (repo list stays the body); pick one with
  `hup tui --mode <name>`. New `multiplex` mode strips the chrome and, on
  selecting a repo, clones it if needed then runs a templated, shell-safe
  `switch_command` (clone-path fields plus `{{.Target}}`) — e.g. jump to a tmux
  session at the checkout.
- `hup tui --flatlist` starts directly in the combined flat repo list; combine
  with `--mode multiplex` to fuzzy-find any repo and jump to it in one command.
- Combined "All repositories" view: an entry in the provider list opens a flat,
  cross-provider list of every cached repo (provider-prefixed rows, a
  completeness badge, and a full cross-provider refresh), with the existing
  fuzzy filter, facets, sort, and clone/detail actions applied across the set.

### Changed

- Multiplex mode is more polished: no multi-select checkbox (single-repo
  action), a successful switch exits cleanly, and cloning-first shows a centered
  progress popup with a real progress bar (cancel with Esc).
- **BREAKING**: `username` is now required for every provider in `config.yaml`
  (`hup config check` reports it if missing). Your own account is an ordinary
  owner named by `username` — pinned first and tinted in the TUI owner list —
  and an empty `owners` list resolves to it. The `self` token in
  `exclude_owners` is gone; exclude your own account by listing your `username`.
- `config.example.yaml` comments normalized to copy-pasteable `# key: value`
  form, and the example is now validated against the config schema in CI so it
  can never drift out of sync.

### Fixed

- The TUI owner view no longer over-shows: selecting your own account now lists
  only your repositories instead of every repo for the provider.
- You can now move the selection while typing a fuzzy filter (↑/↓, PgUp/PgDn,
  Ctrl+N/Ctrl+P); typing narrows and resets to the top, and a single Enter acts
  on the highlighted match instead of requiring two presses.
- The footer (keybindings, position indicator) is now pinned to the bottom of
  the viewport instead of floating up under a short list.

## [1.0.0] - 2026-08-18

### Added

- Multi-provider portfolio manager foundations: YAML config loading/validation,
  configurable clone-path templating, GitHub/Codeberg/GitLab clients, per-provider
  JSON cache, and a cron-friendly CLI sync (clone-missing + fast-forward-pull).
- Owner/org auto-discovery via `all_owners` (with `exclude_owners`); eager sync,
  lazy per-owner fetch in the TUI, over an owner-indexed cache.
- TUI: provider → owner → repo browsing with a scrolling viewport and paging
  keys, a position indicator, level-aware fuzzy filter, facet filters
  (fork/archived/visibility), name/last-updated sorting, a repo detail view with
  a glamour-rendered README, single/bulk clone, and open-in-browser.
- Fetch progress: a shared event stream with a spinner then a determinate bar,
  bounded-parallel page fetching, and cancellation.
- Release process: `VERSION` single source of truth (embedded via `go:embed`,
  read by the flake, overridden by goreleaser from the git tag); goreleaser
  config (linux/darwin × amd64/arm64 with checksums); GitHub Actions release
  workflow on `v*` tags; gated `scripts/release.sh`; `RELEASING.md`; goreleaser
  and gum in the devShell.
- Nix flake (plain, multi-arch) with a coverage gate (`scripts/coverage.sh`,
  ≥70% overall / ≥80% core) enforced by `nix flake check`.

### Changed

- Renamed the project skull2 → golgotha → **HupHop** (binary `hup`); module
  `github.com/mipmip/huphop`, config `~/.config/huphop`, cache `~/.cache/huphop`,
  auth env prefix `HUPHOP_<PROVIDER>_TOKEN`.

### Fixed

- Lazy-fetched owner now displays its repositories immediately, without needing a
  restart (the cache is committed before the completion event is handled).
