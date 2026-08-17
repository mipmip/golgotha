## Context

The TUI browses provider → owner → repos. `visibleRepos()` in
`internal/tui/model.go` builds the repo scope and applies the fuzzy filter,
then returns the slice in fetch/cache order. There is no sort today. Two sort
fields already exist on `provider.Repo`: `Name` and `UpdatedAt`. This change
adds a small sort layer on top of the existing pipeline. Stars sorting is
explicitly out of scope (needs a data-layer field across all providers).

## Goals / Non-Goals

**Goals:**

- Sort the visible repo list by name or last-updated, asc or desc.
- Keep fetch order as the default and a reachable state (non-destructive).
- Compose with the fuzzy filter (sort the filtered subset).
- Match the existing single-key TUI idiom; keep it discoverable via the footer.

**Non-Goals:**

- No stars sorting, no new `Repo` fields, no provider/cache changes.
- No sorting of the provider or owner levels (owners are already alphabetical).
- No persistence of the chosen sort across sessions (resets to fetch order on
  launch).

## Decisions

**Sort state on the Model.** Add two fields: a `sortKey` enum
(`none | name | updated`) and a `sortDir` enum (`asc | desc`). `none` is the
zero value so the default is fetch order with no extra initialization.

**Cycle + reverse keys.** `s` advances `sortKey` through
`none → name → updated → none`. `S` toggles `sortDir`. Chosen over a single
key that folds direction into the cycle because the two-key scheme reaches any
state in fewer presses and reads clearly in the footer (`s: sort  S: reverse`).
`s`/`S` are free in the current keymap (taken: q, j/k, arrows, pg keys,
^u/^d, home/end, `/`, esc, enter, space, c, o, r).

**Sort happens in `visibleRepos()`, after filtering.** The function already
produces the visible slice; appending a `sort.SliceStable` step there means
sort automatically composes with the filter and every scope (owner-level,
provider-level, cross-provider top level). `SliceStable` keeps fetch order as a
stable tiebreaker so equal keys don't jump around.

**Comparison semantics.** Name sorts case-insensitively
(`strings.ToLower`) for intuitive A→Z. Last-updated compares `time.Time`.
`asc`/`desc` flips the comparator result. When `sortKey == none` the function
returns the slice untouched (identical to today).

**Direction defaults per key.** Pressing `s` into a key uses `asc` as the
neutral default; the user presses `S` to get descending. (A per-key smart
default like "updated defaults to newest-first" was considered but rejected as
surprising — one consistent rule is easier to reason about. The footer always
shows the current direction.)

## Risks / Trade-offs

- [Sort interacts with the in-progress `add-repo-filters` change — both edit
  `visibleRepos()`] → Keep the sort step as a clearly separated tail of the
  function (filter first, then sort); coordinate merge order so whichever lands
  second rebases onto the other rather than duplicating the pipeline.
- [Cursor/selection position after a re-sort could feel jumpy] → Re-sorting
  reorders rows under a fixed cursor index; acceptable for an initial version.
  If it feels wrong we can later track the highlighted repo by identity and
  restore its position — noted, not in scope.
- [Case-insensitive name sort allocates lowercased strings per comparison] →
  Negligible for realistic portfolio sizes; optimize only if profiling shows
  it. Not precomputing keeps the change to one function.

## Migration Plan

Pure additive TUI feature; no data or config migration. Rollback is reverting
the single change.

## Open Questions

None — keybindings (`s`/`S`) and default order (fetch order) resolved with the
user before proposing.
