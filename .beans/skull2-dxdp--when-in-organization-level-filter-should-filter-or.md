---
# skull2-dxdp
title: when in organization level, filter should filter orgs
status: completed
type: task
priority: normal
created_at: 2026-08-17T18:56:29Z
updated_at: 2026-08-17T20:48:39Z
parent: skull2-qati
---

## OpenSpec change

Captured as `add-level-aware-filter` (openspec/changes/add-level-aware-filter/) — the fuzzy filter scopes to the current nav level (providers/owners/repos) instead of always flattening to repos; Enter drills the filtered item per level; filter clears on level change; drops top-level global repo search. tui-only, independent. Validated. Ship with: `/ship add-level-aware-filter`.

## Summary of Changes

Shipped via add-level-aware-filter (commit b1cbf3e0). The fuzzy filter now narrows the current nav level (providers / owners / repos), matching owners against the raw name; Enter drills the filtered item per level; filter clears on any level change; top-level global repo search dropped. Overall coverage 80.5%.
