---
# skull2-p22b
title: sorting on alphabet/last-updated
status: todo
type: task
priority: normal
created_at: 2026-08-17T18:15:32Z
updated_at: 2026-08-17T19:34:42Z
parent: skull2-qati
---

Sort the repo list in the TUI by **name (alphabet)** or **last-updated**.
Stars sorting is split out to its own bean (needs a `Stars` data-layer
field across all three providers) — see the sibling stars bean.

## Scope / decisions

- **Keys:** alphabet (`Repo.Name`) and last-updated (`Repo.UpdatedAt`) —
  both already exist on the model. No provider/cache changes needed.
- **Insertion point:** apply the sort in `internal/tui` `visibleRepos()`
  (`model.go`), *after* filtering, so the sort orders the visible subset.
  Today that function returns repos in fetch/cache order (unsorted).
- **UX (to confirm at propose time):** single cycle key (e.g. `s`) to
  rotate sort key, plus a direction toggle for asc/desc. Matches the
  existing one-key TUI idiom.
- **Default order (open):** likely last-updated **desc** (most recently
  active first) or alphabet asc — decide when proposing.
- **Coordinates with `add-repo-filters`** (in-progress): both touch
  `visibleRepos()`; sort runs after the filter step.

Out of scope: stars sorting; sorting the owner list (already alphabetical).
