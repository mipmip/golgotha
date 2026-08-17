---
# skull2-hm82
title: 02d JSON cache & refresh
status: todo
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-6zs4
---

Persist per-provider repo metadata to `~/.cache/skull2/<provider>.json` and expose `skull2 refresh`.

## Tasks
- [ ] Cache schema: fetched_at + []Repo; atomic write, read
- [ ] `skull2 refresh [--provider NAME]` re-fetches and writes cache
- [ ] TUI/sync read from cache as source of truth
- [ ] Unit tests for round-trip and staleness handling
