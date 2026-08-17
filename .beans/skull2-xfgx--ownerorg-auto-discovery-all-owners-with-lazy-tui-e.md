---
# skull2-xfgx
title: Owner/org auto-discovery (all_owners) with lazy TUI + eager sync
status: completed
type: feature
priority: normal
created_at: 2026-08-17T17:37:04Z
updated_at: 2026-08-17T17:55:48Z
parent: skull2-qati
---

Discover all orgs the user belongs to via a single all_owners option; exclude_owners to ignore. Sync eager, TUI lazy per-owner, over a shared owner-indexed cache.

## OpenSpec change

Captured as `add-owner-discovery` (openspec/changes/add-owner-discovery/) — proposal, design and delta specs for configuration, provider-clients, repo-cache, sync and tui, plus tasks. Validated. Ship with: `/ship add-owner-discovery`.

## Summary of Changes

Implemented via add-owner-discovery: single `all_owners` opt-in discovers every org the user belongs to (plus own account via the SelfOwner empty-string sentinel); `exclude_owners` ignores by name (case-insensitive; "self" token). Non-breaking; explicit owners unioned. ListOwners added to all three clients (GitHub/Codeberg /user/orgs, GitLab member groups). Cache v2 gains an owner index with per-owner fetch state and legacy-flat read. Split by command: sync/refresh eager (full backup sweep), TUI lazy (owner list instant, fetch-per-owner on entry, `r` re-fetches). Overall coverage 80.7%. Shipped via /ship (commit 917e1083).
