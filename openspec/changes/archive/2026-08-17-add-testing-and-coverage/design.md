## Context

All features are implemented across milestones 01–04. This change makes the PoC
provably correct and guards it with a coverage gate.

## Goals / Non-Goals

**Goals:**
- Hermetic e2e tests for the refresh→browse→clone and sync flows.
- An enforced coverage gate in the nix checks.

**Non-Goals:**
- Real network/provider integration tests.
- CI configuration beyond `nix flake check`.

## Decisions

- **E2E harness**: `httptest.Server` serving fixture provider JSON + local bare
  git repos as clone sources; everything under `t.TempDir()` with `t.Setenv`
  for HOME/XDG and tokens. Live in an `e2e` package.
- **Coverage measurement**: `go test -coverprofile` across `./...`; a small
  script computes overall and per-core-package percentages and exits non-zero
  below thresholds (overall 70, core 80). Core packages: config, clonepath,
  provider, cache, syncer.
- **Nix wiring**: add a `checks.coverage` derivation that runs the script; keep
  the existing `build`/`gotest` checks.

## Risks / Trade-offs

- [Thresholds can be brittle around the TUI] → TUI excluded from the core-80
  set (smoke/update tests only); overall-70 accommodates it.
- [git binary required for e2e] → available in the nix check environment.
