---
# skull2-x4ej
title: when filtering navigation stops working
status: completed
type: bug
priority: normal
created_at: 2026-08-19T19:51:58Z
updated_at: 2026-08-19T20:35:23Z
parent: skull2-ok4c
---

when using the search filter, moving the selected row is not possible anymore.



## Diagnosis (explore 2026-08-19)

In `handleKey`, while filtering (`m.filtering`), every key except Enter/Esc/CtrlC
is fed to the text input AND `m.cursor`/`m.offset` are reset to 0 on every
keystroke (update.go ~246-250). So arrows do nothing (single-line caret) and the
selection is pinned to row 0 — you can only navigate after pressing Enter to
leave edit mode. The expected UX is fzf-style: type and move at the same time.

## Fix (decided)

Change name: `fix-filter-navigation`. In the filtering block:
- Route navigation keys to `moveCursor` *before* the text input: **Up/Down**
  (±1), **PgUp/PgDn** (±page), and **Ctrl+N/Ctrl+P** (down/up, fzf/emacs). `j`/`k`
  keep typing (letters); Home/End/Ctrl+U stay with the text input for editing.
- Only reset `cursor=0, offset=0` when the query text actually changes (compare
  `filter.Value()` before/after `filter.Update`), so navigation preserves the
  highlight while typing narrows and resets to top.
- **One Enter acts (fzf-style):** while filtering, Enter blurs the input (keeping
  the query) and delegates to the normal level Enter handler so it drills/opens
  the highlighted filtered item in a single press — aligning with the existing
  spec "Enter drills the filtered item at its level". Drilling clears the filter
  as today (filter clears on level change). Update TestFilterEnterDrillsFilteredOwner
  from two Enters to one.

**Touches:** internal/tui/update.go (filtering key block), spec addition
"navigation works while filtering", and the one-Enter test update. tui-only, no
config/schema/gate impact.



## Summary of Changes

Fixed navigation while the fuzzy filter is being typed. In handleKey's filtering
block: Up/Down, PgUp/PgDn, and Ctrl+P/Ctrl+N now move the selection (before the
text input); the cursor resets to the top only when the query text changes
(compare filter.Value() around filter.Update), so navigation preserves the
highlight. Enter now acts in one press (blur + delegate to m.enter()), matching
the spec. j/k/Home/End/Ctrl+U still edit the query. Updated three filter tests to
the one-Enter model and added filter_nav_test.go. Shipped as
2026-08-19-fix-filter-navigation (commit 892b696d).
