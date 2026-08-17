---
# skull2-vle5
title: more filters
status: completed
type: task
priority: normal
created_at: 2026-08-17T18:13:11Z
updated_at: 2026-08-17T20:23:53Z
parent: skull2-qati
---

- forked yes/no
- public/private
- archived yes/no

## OpenSpec change

Captured as `add-repo-filters` (openspec/changes/add-repo-filters/) — Model A narrow-only tri-state facets (fork/archived) + visibility value-cycle, Repo.Visibility string mapped across providers; deltas to provider-abstraction, provider-clients and tui. Validated. Ships AFTER `add-fetch-progress`. Ship with: `/ship add-repo-filters`.

## Summary of Changes

Shipped via add-repo-filters (commit 4adb7cff). Added Repo.Visibility (public/private/internal) mapped across all three providers. Model-A narrow-only TUI facet filters: tri-state fork + archived, visibility value-cycle (f/a/v keys), composed with fuzzy, status line, window reset on change, and a "not cached" hint when an only-facet asks for fetch-excluded data. Overall coverage 81.5%.
