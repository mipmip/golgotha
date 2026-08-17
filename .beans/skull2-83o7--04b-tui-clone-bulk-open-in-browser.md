---
# skull2-83o7
title: 04b TUI clone, bulk & open-in-browser
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:30:34Z
parent: skull2-qa5w
---

Actions on selections: single/bulk clone to templated targets, open in browser, and refresh.

## Tasks
- [x] Single select -> clone to configured target
- [x] Multi-select (space) -> bulk clone with progress
- [x] `o` opens the repo `web_url`
- [x] `r` refreshes current provider cache
- [x] Update-function unit tests for actions

## Summary of Changes

TUI actions: single clone (Enter/c) and multi-select (space) bulk clone via syncer.Engine.CloneRepo, o opens WebURL (stubbable), r refreshes cache. Update pure/testable. Coverage 81%.
