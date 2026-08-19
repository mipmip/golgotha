---
# skull2-qp5y
title: 'TUI: "(your account)" shows all repos, not just your own'
status: in-progress
type: bug
priority: normal
created_at: 2026-08-18T08:46:14Z
updated_at: 2026-08-19T12:42:05Z
parent: skull2-ok4c
---

Self account is keyed by the SelfOwner "" sentinel; its repos carry their real owner login, so the self view over-shows.

## Diagnosis

The self account is represented by the `SelfOwner` sentinel `""`, but the repos it fetches carry their real owner login (e.g. `mipmip`). Two consequences in the TUI:

- `visibleRepos()` treats `selOwner == ""` as "no owner selected" and falls through to the default branch that shows ALL provider repos.
- `cache.ReposFor("")` never matches (repos have a real owner), so reload/scope by the sentinel yields the wrong set.

Result: selecting "(your account)" shows every repo across all owners, not just your own.

## Agreed fix (mandatory `username` config)

Rather than resolving the self login over the network, add a **mandatory
`username` field per provider** in config. This is known offline at config
load, so it removes the whole request/caching path *and* lets us **delete the
`SelfOwner` sentinel entirely**. (Change name: `fix-self-owner-resolution`.)

The sentinel `""` only ever existed to answer one question at fetch time: "is
this owner me (→ `/user/repos`, private included) or an org (→
`/orgs/<x>/repos`)?" With a configured username the client just compares
`owner == cfg.Username` — no `GET /user`, no cache field, no cold-start
placeholder, no `""` guards.

### Before → after

| Concern              | Before (sentinel)                | After (mandatory username)            |
|----------------------|----------------------------------|---------------------------------------|
| config owners        | `["", "acme", "beta"]`           | `["mipmip", "acme", "beta"]`          |
| resolve self id      | (would need GET /user + cache)   | `cfg.Username` (offline, free)        |
| fetch routing        | `owner==""` → `/user/repos`      | `owner==cfg.Username` → `/user/repos` |
| `cache.ReposFor`     | miss on `""` ✗ (the bug)         | match on `"mipmip"` ✓                 |
| `visibleRepos` scope | `""` falsy → shows all ✗         | `"mipmip"` scopes correctly ✓         |
| TUI label            | `ownerLabel: "" → "(your acct)"` | real login, tinted                    |
| TUI pin/tint         | special-case `""`                | special-case `owner==cfg.Username`    |
| cold start           | placeholder until first fetch    | correct from load                     |

### Scope (decided with pim)

- **Bug fix + tint only.** Resolve self via mandatory `username`; self behaves
  like an ordinary owner but is **pinned first** and **tinted** in the owner
  list. Label is the real login (no `(your account)` string).
- Legacy "empty `owners:` = my own repos" still works — it resolves to
  `[username]` instead of `[""]`.
- `username` required-field check rides in this change's tasks (extends the
  existing `Config.Validate()`); config validation framework already exists
  (`skull2-0s31`), so no separate validation bean.

### Costs (accepted)

- Breaking config change: `username` becomes required per provider (clear load
  error if missing). BRIEFING.md + `openspec/config.yaml` schema updated.
- We trust the typed value — no network verifies it matches the token's real
  identity. Load-time check is non-empty only.

### Deferred (not this bean)

- pim's wish for a **combined cross-provider "all repos" list** is a genuinely
  new navigation mode (nav today is strictly Providers→Owners→Repos). It
  belongs with `skull2-wzbf` (tui-modes / multiplexer), not this bug fix.
