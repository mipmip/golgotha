---
# skull2-q75f
title: shortcut enter should show repo details
status: completed
type: task
priority: normal
created_at: 2026-08-17T18:10:38Z
updated_at: 2026-08-17T20:37:17Z
parent: skull2-qati
---

- Show readme with `bat` if available,
- topics, stars, last updated, description etc...
- Lazy fetch to cache at open view.

## OpenSpec change

Captured as `add-repo-details` (openspec/changes/add-repo-details/) — Enter opens a repo detail view (metadata + glamour-rendered README), lazy fetch of stars/topics/language + README cached in per-repo detail files (not the big json), graceful offline. New capability repo-details + deltas to provider-abstraction, provider-clients, tui. NOTE: uses glamour (new dep, vendorHash bump), superseding the bean's original bat idea. Ships AFTER add-repo-filters. Ship with: `/ship add-repo-details`.

## Summary of Changes

Shipped via add-repo-details (commit fd877157). Repo-level Enter opens a detail view (metadata + glamour-rendered scrollable README); c is the sole clone key. Added RepoDetails + Readme to the provider interface, implemented for all three providers. Lazy fetch on first open, cached in per-repo files under ~/.cache/skull2/details (separate from list cache); r re-fetches; graceful offline. Added glamour dep (vendorHash bumped). Overall coverage 80.2%.
