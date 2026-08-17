## Why

The core backup use case is a headless, cron-friendly command that keeps local
clones in step with the providers: clone what is missing and fast-forward-pull
what exists, safely.

## What Changes

- Implement a sync engine that, for each cached repository, resolves its target
  path via the clone-path engine, clones it if absent, and otherwise fetches +
  fast-forward-only pulls the default branch — never forcing, skipping dirty
  trees with a warning.
- Produce a structured per-provider summary (cloned / updated / skipped /
  failed).
- Add the `skull2 sync [--provider NAME] [--no-refresh]` command: refresh the
  cache first (unless `--no-refresh`), then run the engine; line-oriented logs
  and non-zero exit on any failure.

## Capabilities

### New Capabilities
- `sync`: the clone-missing + fast-forward-pull engine and the `skull2 sync`
  command.

### Modified Capabilities
<!-- none -->

## Impact

- `internal/syncer` (new implementation).
- `cmd/skull2` (new `sync` subcommand).
- Uses the local `git` binary; no new Go dependencies expected.
