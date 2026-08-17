---
# skull2-q75f
title: shortcut enter should show repo details
status: todo
type: task
priority: normal
created_at: 2026-08-17T18:10:38Z
updated_at: 2026-08-17T19:15:27Z
parent: skull2-qati
---

- Show readme with `bat` if available,
- topics, stars, last updated, description etc...
- Lazy fetch to cache at open view.

## OpenSpec change

Captured as `add-repo-details` (openspec/changes/add-repo-details/) — Enter opens a repo detail view (metadata + glamour-rendered README), lazy fetch of stars/topics/language + README cached in per-repo detail files (not the big json), graceful offline. New capability repo-details + deltas to provider-abstraction, provider-clients, tui. NOTE: uses glamour (new dep, vendorHash bump), superseding the bean's original bat idea. Ships AFTER add-repo-filters. Ship with: `/ship add-repo-details`.
