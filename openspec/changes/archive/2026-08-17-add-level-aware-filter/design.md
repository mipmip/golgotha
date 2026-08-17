## Context

`bodyText` renders `renderRepos(visibleRepos())` whenever the filter is active,
ignoring `m.nav`; `enter()` has a branch that "operates on the flattened repo
view directly" while filtering. So `/` is always a repo search. This change
scopes the filter to the current level.

## Goals / Non-Goals

**Goals:**
- Filter narrows the current level's list (providers / owners / repos).
- Enter drills the highlighted filtered item per level.
- Predictable filter lifecycle across navigation.

**Non-Goals:**
- A separate global repo search (could be a future dedicated key).
- Any change to fetching, data model, or providers.

## Decisions

- **Level-scoped filtering:** in `bodyText`, when the filter is active, filter
  the current level's items with the existing `fuzzyMatch` helper:
  - providers → `providerNames()`
  - owners → owner labels from `ownersFor(selProvider)`
  - repos → `visibleRepos()` (unchanged)
  Factor a small `filtered(items)` helper (or per-level `visibleProviders` /
  `visibleOwners`) so windowing/rendering stay uniform.
- **Enter under filter:** remove the flatten shortcut in `enter()`; use the
  normal per-level branch against the filtered list (drill into the highlighted
  provider/owner, or clone/open a repo).
- **Filter lifecycle:** clear the filter on any level change (drill in via
  `enter`, and back via `goBack` — which already clears it first). Each level
  starts unfiltered; cursor/offset reset as they do today.
- **Drop global repo search from the top:** acceptable because with lazy
  `all_owners` it only matched already-fetched repos.

## Risks / Trade-offs

- [Loss of global repo search] → low value under lazy fetch; revisit later with a
  dedicated key if wanted.
- [Owner labels include decorations like "(not fetched)"] → fuzzy-match against
  the raw owner name, not the decorated label, so matching stays predictable.
