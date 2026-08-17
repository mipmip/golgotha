---
# skull2-hm82
title: 02d JSON cache & refresh
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:13:59Z
parent: skull2-6zs4
---

Persist per-provider repo metadata to `~/.cache/skull2/<provider>.json` and expose `skull2 refresh`.

## Tasks
- [x] Cache schema: fetched_at + []Repo; atomic write, read
- [x] `skull2 refresh [--provider NAME]` re-fetches and writes cache
- [x] TUI/sync read from cache as source of truth
- [x] Unit tests for round-trip and staleness handling

## Summary of Changes

internal/cache: per-provider JSON at ~/.cache/skull2 (XDG aware), atomic Save (temp+rename), Load/LoadOrEmpty; skull2 refresh [--provider]. Coverage 80%.
