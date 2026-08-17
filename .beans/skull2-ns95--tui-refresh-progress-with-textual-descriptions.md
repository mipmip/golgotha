---
# skull2-ns95
title: 'TUI: refresh progress with textual descriptions'
status: in-progress
type: feature
priority: normal
created_at: 2026-08-17T18:00:44Z
updated_at: 2026-08-17T19:50:22Z
parent: skull2-qati
---

With a lot of organizations it takes long to refresh. there should be a progress bar showing what the app is doing

## OpenSpec change

Captured as `add-fetch-progress` (openspec/changes/add-fetch-progress/) — one progress-event model (new fetch-progress capability) with deltas to provider-clients, tui, sync and repo-cache; bounded-parallel page/owner fetch (cap 6), cancel-current-fetch, commit-only-on-complete. Validated. Ship with: `/ship add-fetch-progress`.
