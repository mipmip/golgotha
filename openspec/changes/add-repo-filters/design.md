## Context

Filtering today is fetch-time (`FilterRepos` drops archived/forks before the
cache) and the `Repo` model has no visibility field. The fuzzy `/` filter is the
only view-time filter. This change adds interactive facet filtering in the TUI
over the cached set (Model A, narrow-only) and captures visibility.

## Goals / Non-Goals

**Goals:**
- Interactive fork / archived / visibility filtering while browsing.
- Capture repository visibility across providers.
- No behavior change until a filter is touched.

**Non-Goals:**
- Caching everything / moving fetch-time filtering to view-time (that was Model
  B, not chosen).
- New config keys or a sync-time filter.
- Persisting filter state between runs (in-memory for the session).

## Decisions

- **Model A (narrow-only):** filters subset whatever is cached. The cache
  superset is still governed by config `include_archived` / `include_forks`.
  Consequence: the archived/fork facets can only reveal what config already
  cached — surface a hint when a facet would match excluded-at-fetch data.
- **Visibility as a string** (`public` / `private` / `internal`) for headroom
  (GitHub-EE / GitLab `internal`). Map: GitHub `private` bool → public/private;
  Gitea `private` bool → public/private; GitLab `visibility` passed through.
  Unknown/empty normalizes to `public`.
- **Facet model:**
  - fork, archived: tri-state `all · only · hide`.
  - visibility: value-cycle `all · public · private · internal`.
  - Facets AND together, then the fuzzy query narrows further.
- **Defaults mirror config:** facets start at `all`, except fork starts at
  `hide` when `include_forks:false` and archived is effectively `hide` already
  because they aren't cached — keeping current behavior until toggled.
- **UX:** key-toggled filter bar — e.g. `f` cycles fork, `a` archived, `v`
  visibility; a status line shows active facets (`fork:hide vis:private`).
  Filtering is applied in the TUI's `visibleRepos` pipeline alongside fuzzy.
- **State:** in-memory on the model; resets are fine on quit. Windowing/paging
  (already present) recompute against the filtered list; offset resets when the
  filter set changes (same rule as fuzzy keystrokes).

## Risks / Trade-offs

- [Conditional facets confuse users] → explicit hint when a facet matches
  nothing because the data was excluded at fetch.
- [Overlap with add-fetch-progress in client mappings] → sequence after it;
  add `Visibility` against the refactored fetch to avoid conflicts.
- [Visibility scope] → private repos only appear if the token can see them;
  that's pre-existing and out of scope here.
