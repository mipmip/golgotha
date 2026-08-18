---
# skull2-qp5y
title: 'TUI: "(your account)" shows all repos, not just your own'
status: todo
type: bug
priority: normal
created_at: 2026-08-18T08:46:14Z
updated_at: 2026-08-18T08:46:14Z
parent: skull2-qati
---

Self account is keyed by the SelfOwner "" sentinel; its repos carry their real owner login, so the self view over-shows.

## Diagnosis

The self account is represented by the `SelfOwner` sentinel `""`, but the repos it fetches carry their real owner login (e.g. `mipmip`). Two consequences in the TUI:

- `visibleRepos()` treats `selOwner == ""` as "no owner selected" and falls through to the default branch that shows ALL provider repos.
- `cache.ReposFor("")` never matches (repos have a real owner), so reload/scope by the sentinel yields the wrong set.

Result: selecting "(your account)" shows every repo across all owners, not just your own.

## Proposed fix

Resolve `SelfOwner` to the authenticated user's real login at fetch time (one `GET /user` per provider) and key everything — owner index, `ReposFor`, `visibleRepos` scope, and display — by that real login. Then "your account" is just an owner like any other; no sentinel, no mismatch. (Change name: `fix-self-owner-resolution`.)
