---
# skull2-4pob
title: add paging when lists are too long e.g. 200+ repo's.
status: completed
type: feature
priority: normal
created_at: 2026-08-17T14:37:02Z
updated_at: 2026-08-17T17:11:31Z
parent: skull2-qati
---

## OpenSpec change

Captured as `add-tui-list-scrolling` (openspec/changes/add-tui-list-scrolling/) — proposal, design, tui spec delta and tasks authored and validated. Ready for `/opsx:apply`.

Ship with: `/ship add-tui-list-scrolling` (see [[automate-single-change-shipping-ship]]).

## Summary of Changes

Implemented via add-tui-list-scrolling: hand-rolled scrolling viewport on all three list levels (offset keeps cursor visible, fixed scroll-off margin 2, height<=0 renders all), PgUp/PgDn + Ctrl-U/Ctrl-D + Home/End keys, and a 'first-last of n' position indicator. tui coverage 82.8%. Shipped via /ship (commit 7998e9b3).
