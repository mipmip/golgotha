## Context

The self account is currently represented by the `SelfOwner` sentinel
(`config.SelfOwner == ""`). Providers fetch it via the authenticated endpoint
(`/user/repos`), but the returned repos carry their **real** owner login (e.g.
`mipmip`). The sentinel is therefore never reconcilable with the data:

- `cache.ReposFor("")` matches nothing (repos are keyed `mipmip`).
- `visibleRepos()` reads `selOwner == ""` as "no owner selected" and falls
  through to showing the whole provider — the bug in `skull2-qp5y`.

The gap exists because the self login is unknown at config-load time. The fix
closes that gap by making the login a **mandatory config value** (`username`),
which is known offline and lets the sentinel — and the whole class of
sentinel/real-login mismatch bugs — be deleted.

## Goals / Non-Goals

**Goals:**

- Stop the self view over-showing; entering the self owner shows only the user's
  own repositories.
- Delete the `SelfOwner` sentinel and every `""` special-case across config,
  providers, cache, TUI and CLI.
- Make the self account an ordinary owner, pinned first and tinted in the TUI.
- Resolve identity with **zero** network calls and **zero** caching.

**Non-Goals:**

- No `GET /user` / identity discovery: identity comes from config.
- No combined cross-provider "all repos" view (deferred to `skull2-wzbf`).
- No change to auth resolution, pagination, sync, or the clone-path engine.

## Decisions

### Decision: Mandatory `username` per provider, sentinel deleted

Add a required `username string \`yaml:"username"\`` to `config.Provider`.
`Config.Validate()` rejects an empty `username` with an actionable, per-provider
error alongside the existing `name`/`type`/`short` checks. `config.SelfOwner`
and `selfExcludeToken` and all `owner == SelfOwner` / `owner == ""` branches are
removed.

- **Why:** identity is knowable offline; a configured value removes the only
  reason the sentinel existed.
- **Alternative — resolve at fetch via `GET /user` + cache:** authoritative and
  needs no config change, but adds a network round-trip, a persisted identity
  field, a cold-start placeholder, and keeps the sentinel alive. Rejected as
  more code for a problem the config value dissolves. Speed matters; this path
  makes zero extra requests.
- **Alternative — derive login from the fetched repos:** breaks for a provider
  where the user owns zero repos (no repo → no login). Rejected.

### Decision: Fetch routing by `owner == cfg.Username`

Each provider client already holds `cfg`. `FetchOwner(owner)` chooses the
endpoint by comparing `owner` to `cfg.Username`:

- `owner == cfg.Username` → authenticated endpoint (`/user/repos?affiliation=owner`
  on GitHub; the equivalent authenticated-user listing on GitLab/Forgejo),
  which includes private repos.
- otherwise → the organization/group endpoint for `owner`.

This is the same distinction the sentinel used to make, now keyed on a real
name. `ListOwners` (org discovery via `/user/orgs` etc.) is unchanged.

### Decision: `ResolveOwners` keys self on `username`

`ResolveOwners` no longer prepends the `""` sentinel. Instead:

- `all_owners: false` with empty `owners:` → `[username]` (was `[SelfOwner]`).
- `all_owners: true` → union of `username`, discovered orgs, explicit `owners`,
  minus `exclude_owners`. `exclude_owners` matches the `username`
  case-insensitively (the dedicated `"self"` token is dropped — exclude by the
  real name).

Ordering for the config layer is not load-bearing; the TUI owns display order.

### Decision: TUI pins + tints the self owner

The TUI already has the provider config, so it knows `username`. In
`ownersFor(provider)` the owner whose name equals `provider.Username` is pulled
out and prepended (pinned first); the rest sort as today. The renderer applies a
distinct lipgloss style to that row. `ownerLabel` loses its `SelfOwner` →
`"(your account)"` mapping and simply shows the real login. `visibleRepos()`
loses the `selOwner != ""` guard; the selected owner is always a real login, so
`it.Repo.Owner == m.selOwner` scopes correctly — including for the self owner.

### Decision: Cache and CLI drop sentinel handling

`internal/cache` keys owners by real login already; only the documented
"`SelfOwner` sentinel is supported" note and any `""` test fixtures are removed.
`cmd/hup` drops `displayOwner`/`SelfOwner` and defaults the owner set to
`[username]`.

## Risks / Trade-offs

- **[Breaking config change]** → Existing `config.yaml` files without `username`
  fail to load. Mitigation: a clear, per-provider validation error; update
  BRIEFING.md and `openspec/config.yaml` and the example config so the new
  required field is documented. Acceptable at v1.0.0 with a single known user.
- **[Typo'd username is unverified]** → If `username` does not match the token's
  real identity, the self fetch still hits `/user/repos` (the token's repos) but
  labels them under the typo'd name; org rows for the real login would 404/empty.
  Mitigation: it is the user's own config; validation guarantees non-empty but
  cannot verify correctness without the network call we are deliberately
  avoiding. Documented behavior.
- **[Redundant when a provider wants no self repos]** → `username` is still
  required even if the user only lists orgs. Accepted for a single, simple rule;
  the value is cheap and the future combined view will want it anyway.

## Migration Plan

1. Users add `username: <login>` under each provider in `config.yaml`.
2. `hup config check` reports the missing field until it is added.
3. No cache migration needed — self repos were already stored under their real
   login; removing the sentinel only fixes lookups that previously missed.
