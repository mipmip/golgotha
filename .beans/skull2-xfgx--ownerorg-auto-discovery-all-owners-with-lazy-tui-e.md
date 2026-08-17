---
# skull2-xfgx
title: Owner/org auto-discovery (all_owners) with lazy TUI + eager sync
status: in-progress
type: feature
priority: normal
created_at: 2026-08-17T17:37:04Z
updated_at: 2026-08-17T17:40:41Z
parent: skull2-qati
---

Discover all orgs the user belongs to via a single all_owners option; exclude_owners to ignore. Sync eager, TUI lazy per-owner, over a shared owner-indexed cache.

## OpenSpec change

Captured as `add-owner-discovery` (openspec/changes/add-owner-discovery/) — proposal, design and delta specs for configuration, provider-clients, repo-cache, sync and tui, plus tasks. Validated. Ship with: `/ship add-owner-discovery`.
