## Context

The cache holds repositories and the clone-path engine resolves targets. This
change adds the engine that reconciles the local filesystem with the cache.

## Goals / Non-Goals

**Goals:**
- Safe, idempotent clone-missing + ff-pull-existing across providers.
- Cron-friendly output and exit codes.

**Non-Goals:**
- Mirror/bare backups (BRIEFING chose working-tree sync).
- Conflict resolution beyond skipping dirty trees.

## Decisions

- **Git via the `git` binary**: shell out to `git` (clone, fetch, `merge
  --ff-only`, `status --porcelain`, `rev-parse`) rather than a pure-Go git
  library, to match real credential/SSH behavior. Wrap it behind a small
  interface so tests use temp git repos (or a fake runner).
- **Dirty detection**: `git status --porcelain` non-empty → skip + warn.
- **Default branch**: use the repo's recorded default branch; fall back to the
  remote HEAD.
- **Concurrency**: sequential per provider for the PoC; summary aggregated in a
  struct. Failures are collected, not fatal mid-run; the command exits non-zero
  if any failed.

## Risks / Trace-offs

- [Shelling to git couples tests to a git binary] → tests create real temp
  repos with `git init`; the binary is available in the dev shell and nix
  check.
- [SSH auth in cron] → documented; relies on the user's ssh-agent/keys, same as
  manual clones.
