## Why

Selecting "(your account)" in the TUI shows **every** repository for a provider,
not just the user's own (bean `skull2-qp5y`). The self account is keyed by the
`SelfOwner` sentinel `""`, but the repos it fetches carry their real owner login
(e.g. `mipmip`), so `cache.ReposFor("")` misses and `visibleRepos()` reads
`selOwner == ""` as "nothing selected" and falls through to the whole provider.

The root cause is that the self login is unknown at config-load time, so the
code uses a sentinel it can never reconcile with the real login on the repos.
Making the login a **mandatory config value** removes that gap entirely —
identity is known offline, so the sentinel and its whole class of mismatch bugs
can be deleted.

## What Changes

- **BREAKING**: add a **required** per-provider `username` field to
  `config.yaml`. Config load fails with an actionable error when it is missing.
- **Delete the `SelfOwner` sentinel** (`config.SelfOwner`, the `""` value and
  all the guards that special-case it) across config, providers, cache, TUI and
  CLI. The user's account becomes an ordinary owner named by `username`.
- Owner resolution keys the self account on `username`: an empty `owners:` list
  resolves to `[username]`, and `all_owners` unions `username` with discovered
  orgs (instead of the sentinel).
- Provider clients route the self fetch by comparing `owner == cfg.Username`
  (→ authenticated `/user/repos`, private repos included) versus an org
  (→ `/orgs/<owner>/repos`). No `GET /user`, no identity caching.
- TUI: the self owner behaves like any other owner but is **pinned first** and
  **tinted** in the owner list; its label is the real login (the
  `"(your account)"` string is removed). `visibleRepos()` scopes correctly
  because the selected owner now matches the repos' real login.
- Extend `Config.Validate()` to require a non-empty `username` per provider.
- Update `BRIEFING.md` and `openspec/config.yaml` `context:` for the new
  mandatory field.

Out of scope (deferred to `skull2-wzbf`): a combined cross-provider "all repos"
view. Today's navigation is strictly provider → owner → repos; a flat
cross-provider list is a new mode, not part of this bug fix.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `configuration`: add the mandatory `username` field and its validation; owner
  resolution keys the self account on `username` instead of the `SelfOwner`
  sentinel (empty `owners:` → `[username]`; `all_owners` unions `username`).
- `provider-clients`: the self-vs-org fetch endpoint is chosen by comparing the
  requested owner to the configured `username`, not by an empty sentinel owner.
- `tui`: the owner list shows the self account under its real login, pinned
  first and visually distinguished; repository scoping for the self owner
  matches its real login.

## Impact

- **Config schema (BREAKING)**: `internal/config` — new required `username`
  field, `Validate()`, and `ResolveOwners` (drops the sentinel). Existing
  configs without `username` fail to load until updated.
- **Providers**: `internal/provider` — `github.go`, `gitlab.go`, `codeberg.go`
  fetch routing keyed on `cfg.Username`; remove `SelfOwner` handling.
- **Cache**: `internal/cache` — owner keys become real logins; drop the
  documented `SelfOwner` support note.
- **TUI**: `internal/tui` — owner list pin + tint, remove `ownerLabel`'s
  `"(your account)"` mapping, fix `visibleRepos()` self scoping.
- **CLI**: `cmd/hup` — remove `displayOwner`/`SelfOwner` defaults; default owner
  set becomes `[username]`.
- **Docs**: `BRIEFING.md`, `openspec/config.yaml`.
- Tests across the above packages; no new dependencies.
