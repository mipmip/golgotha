## Why

Today you must name every organization in `owners:` to see its repos, and on
GitHub/Codeberg listing `owners` even drops your own repos. Managing a large
portfolio means constantly editing config as org membership changes. Skull2
should be able to discover every org you belong to automatically, while still
letting you ignore specific ones.

## What Changes

- Add a single opt-in option `all_owners` that discovers every organization the
  authenticated user belongs to (plus the user's own account), and
  `exclude_owners` to ignore specific ones by declaration. Non-breaking: default
  off; explicit `owners` keeps working and is unioned in.
- Add owner/org discovery to each provider client (GitHub `/user/orgs`, Codeberg
  `/user/orgs`, GitLab member groups).
- Split fetching by command over a shared cache:
  - `skull2 sync` / `skull2 refresh` is **eager** — discover, then fetch repos
    for every (non-excluded) owner so backups and the cache cover everything.
  - the TUI is **lazy** — show the discovered owner list immediately and fetch a
    given owner's repos the first time you enter it, caching the result.
- Extend the cache with an owner index and per-owner fetch state so the TUI can
  list owners before their repos exist and know which still need fetching.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `configuration`: add `all_owners` (bool) and `exclude_owners` ([]string); define
  owner-set resolution.
- `provider-clients`: add owner/org discovery per provider.
- `repo-cache`: add an owner index + per-owner fetch state (discovered vs fetched).
- `sync`: eager discovery sweep across all resolved owners.
- `tui`: owner level sourced from discovered owners; lazy per-owner repo fetch
  with a loading state.

## Impact

- `internal/config`, `internal/provider`, `internal/cache`, `internal/syncer`,
  `internal/tui`, and the `refresh`/`sync` commands. No new dependencies expected.
- Requires appropriate token scope for discovery (e.g. GitHub `read:org` for
  private org membership).
