## Why

The TUI navigation is strictly hierarchical (provider → owner → repos), so there
is no way to see a single combined list of every repository across all providers
and owners at once (bean `skull2-39es`, split off from `skull2-qp5y`). Users want
a "portfolio overview": one flat, filterable, sortable list of everything, from
which they can open or clone any repo regardless of which provider or owner it
belongs to.

## What Changes

- Add a synthetic **"All repositories"** entry pinned at the top of the provider
  list. Selecting it opens a **flat, cross-provider repository view**.
- The flat view reuses the existing repos level with a `flatAll` scope flag
  (no new navigation level type): `visibleRepos()` takes its existing
  all-providers branch, ignoring the drilled-down provider/owner selection.
- Rows are disambiguated with a **provider-short + owner/name** prefix
  (e.g. `gh  mipmip/foo`). The existing fuzzy filter, facet filters, sort, single
  and bulk clone, and detail view all operate on the flat list unchanged.
- Show a **completeness badge** (e.g. `147 repos · 3/8 owners loaded`) computed
  from the owner index versus fetched-owner state, so partial caches are honest
  rather than silently incomplete.
- Provide a **full refresh** action in the flat view that re-fetches every owner
  across every provider (freshness), reusing the existing eager-fetch machinery,
  with progress feedback.
- Leaving the flat view (Esc) returns to the provider list.

Out of scope (own beans): configurable columns (`skull2-n3i2`), sort-by-stars
(`skull2-2h8p`), large-list paging keys (`skull2-4pob`), the modes framework
(`skull2-wzbf`), and **self-account tinting** — tint needs the resolved
`username` from `fix-self-owner-resolution` and is deferred to a follow-up so
this change stays independent of the config-schema work.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `tui`: add a combined, cross-provider flat repository view reachable from a
  top-level "All repositories" entry, with a completeness badge and a full
  cross-provider refresh; existing browsing, filtering, sorting and actions are
  unchanged.

## Impact

- **TUI**: `internal/tui` — synthetic top-level entry at the providers level; a
  `flatAll` scope flag routing to the existing repos rendering; provider-prefixed
  row rendering; completeness-badge computation; a full refresh-all command
  across providers×owners with progress.
- **No config schema change** — independent of `fix-self-owner-resolution`,
  `configurable-tui-chrome`, and `add-config-example-gate`.
- **Follow-up**: self-account tint in the flat view once `fix-self-owner-resolution`
  lands (`username` available).
- **Tests**: TUI unit tests for entry/scope, flat aggregation, prefixing, badge
  counts, and refresh-all orchestration. No new dependencies.
