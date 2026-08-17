---
# skull2-ns95
title: 'TUI: refresh progress with textual descriptions'
status: completed
type: feature
priority: normal
created_at: 2026-08-17T18:00:44Z
updated_at: 2026-08-17T20:16:15Z
parent: skull2-qati
---

With a lot of organizations it takes long to refresh. there should be a progress bar showing what the app is doing

## OpenSpec change

Captured as `add-fetch-progress` (openspec/changes/add-fetch-progress/) — one progress-event model (new fetch-progress capability) with deltas to provider-clients, tui, sync and repo-cache; bounded-parallel page/owner fetch (cap 6), cancel-current-fetch, commit-only-on-complete. Validated. Ship with: `/ship add-fetch-progress`.

## Summary of Changes

Shipped via add-fetch-progress (commit 50557d1d). New internal/fetch package: event model (Started/PageDone/Done/Failed/Canceled/Warning), bounded worker pool (cap 6), generic Pages orchestrator. Providers gained page-aware FetchOwner (total via Link rel=last / X-Total-Pages / X-Total-Count) with ListRepos as a thin wrapper. TUI streams events (spinner→bar + per-page line), Esc cancels; CLI prints per-owner progress, owners fetched bounded-parallel. Cache commits an owner only on complete fetch. Overall coverage 81.0%.
