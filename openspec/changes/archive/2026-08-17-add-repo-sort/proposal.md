## Why

The TUI shows repositories in whatever order the provider API/cache returned —
effectively arbitrary. Users browsing a large portfolio want to impose order:
find a repo by name, or see what changed most recently. Sorting by name and by
last-updated needs no new data (both fields already exist on the `Repo`
model), so it is a small, high-value TUI addition.

## What Changes

- Add a repository-list sort with two keys: **name** (alphabetical) and
  **last-updated** (`UpdatedAt`).
- Bind `s` to cycle the sort key and `S` (shift-s) to reverse direction
  (ascending/descending).
- Keep the current fetch/cache order as the default ("unsorted") and as a
  reachable state in the cycle: `none → name → updated → none`.
- Apply the sort in the TUI *after* filtering, so it orders exactly the
  visible subset (composes with the existing fuzzy filter).
- Surface the active sort in the footer help bar.

Out of scope: sorting by stars (tracked separately — requires a `Stars`
data-layer field across all three providers), and sorting the owner list
(already alphabetical).

## Capabilities

### New Capabilities
<!-- none: sorting extends the existing tui capability -->

### Modified Capabilities
- `tui`: add a requirement for repository-list sorting (sort keys, cycle/
  reverse keybindings, interaction with the fuzzy filter, footer legend).

## Impact

- **Code**: `internal/tui/model.go` (`visibleRepos()` gains a sort step plus
  sort-key/direction state on `Model`); `internal/tui/update.go` (`s` / `S`
  key handling); `internal/tui/view.go` (`footerText` legend). New unit tests
  for sort ordering, direction toggle, cycle wrap-around, and sort-after-
  filter composition.
- **Data**: none — sorts on existing `Repo.Name` and `Repo.UpdatedAt`.
- **Coordination**: `add-repo-filters` (in-progress) also touches
  `visibleRepos()`; the sort step runs after the filter step.
