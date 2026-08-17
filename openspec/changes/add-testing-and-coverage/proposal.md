## Why

The PoC must be provably correct. This change hardens unit coverage, adds
end-to-end tests against mocked provider APIs and temporary directories, and
enforces a coverage gate in `nix flake check` so regressions fail the build.

## What Changes

- Raise unit coverage on core-logic packages (config, clonepath, provider,
  cache, syncer) to at least 80%.
- Add end-to-end tests: mock provider HTTP servers and fixture repos proving
  refresh → browse → clone (to the correct templated path) and sync
  (clone-missing then fast-forward-pull), runnable headlessly under nix.
- Add a coverage gate: overall ≥70% and core packages ≥80%, enforced so a
  failing threshold fails `nix flake check`.

## Capabilities

### New Capabilities
- `quality-gate`: the coverage measurement and threshold enforcement plus the
  end-to-end test harness.

### Modified Capabilities
<!-- none -->

## Impact

- Test files across `internal/...` and a top-level `e2e` test package/dir.
- `flake.nix` `checks` (coverage gate), and a coverage script.
