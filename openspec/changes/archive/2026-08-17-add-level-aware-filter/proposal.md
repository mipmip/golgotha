## Why

The fuzzy filter always flattens to a repository search, regardless of
navigation level. At the owner/organization level this abandons the org list
and drops the user into a repo view — surprising and unhelpful. The filter
should narrow whatever level you are on.

## What Changes

- Make the fuzzy filter **level-aware**: it filters the current level's items —
  provider names at the providers level, owner/org names at the owners level,
  and repositories at the repos level.
- `enter()` no longer takes a "flatten to repos" shortcut while filtering; it
  drills the highlighted (filtered) item using the normal per-level logic.
- Clear the active filter on any level change (drill in or back), so each level
  starts unfiltered.
- Drop top-level global repo search (it only ever matched already-fetched repos
  under the lazy `all_owners` model).

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `tui`: the fuzzy filter scopes to the current navigation level.

## Impact

- `internal/tui` only (`bodyText`, `enter`, filter lifecycle). Windowing already
  wraps the rendered list, so filtered lists scroll for free. No data, provider,
  or dependency changes. Independent of the other queued changes.
