## Why

Browsing a large portfolio needs quick, interactive narrowing beyond the name
fuzzy filter. Users want to slice by fork status, visibility (public/private),
and archived status while browsing. Today archived/fork filtering happens only
at fetch time via config, and visibility isn't captured at all.

## What Changes

- Add a `Visibility` string field to the `Repo` model (public / private /
  internal) and map it from each provider (GitHub `private`, GitLab
  `visibility`, Gitea `private`).
- Add interactive tri-state facet filters to the TUI, composing with the
  existing fuzzy filter:
  - fork: all · only · hide
  - archived: all · only · hide
  - visibility: all · public · private · internal (value cycle)
- Filters are **narrow-only (Model A)**: they subset whatever is already cached.
  Defaults mirror config (start at `all`, or fork=hide when
  `include_forks:false`), so behavior is unchanged until a facet is touched.
- When a facet can't match because the data was excluded at fetch (e.g. "only
  archived" with `include_archived:false`), the TUI shows a hint rather than an
  empty-looking list.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `provider-abstraction`: add `Repo.Visibility`.
- `provider-clients`: populate `Visibility` in the GitHub, Codeberg and GitLab
  mappings.
- `tui`: interactive tri-state facet filters with hints, composed with fuzzy.

## Impact

- `internal/provider` (Repo model + client mappings), `internal/tui` (filter
  state + view). No new dependencies. No config schema change (existing
  `include_archived`/`include_forks` remain the cache superset and default
  facet state).
- Sequenced AFTER `add-fetch-progress`, which refactors the same client mapping
  code; write the client deltas against the refactored fetch.
