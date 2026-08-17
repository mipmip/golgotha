---
# skull2-bbww
title: 04a TUI browse & fuzzy filter
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:30:34Z
parent: skull2-qa5w
---

Bubble Tea UI to navigate provider -> owner -> repos with a global fuzzy filter, reading from the JSON cache.

## Tasks
- [x] Hierarchic navigation model (provider -> owner -> repos)
- [x] Global fuzzy filter (`/`) across flattened repos
- [x] Already-cloned vs not indicator; help/footer bar; `q` quits
- [x] Update-function unit tests

## Summary of Changes

internal/tui: Bubble Tea root model with provider->owner->repo navigation reading the cache, global fuzzy filter (/), cloned indicator, help footer, q to quit.
