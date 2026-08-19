---
# skull2-2h8p
title: sort repos by stars
status: draft
type: task
priority: normal
created_at: 2026-08-17T19:34:33Z
updated_at: 2026-08-18T09:29:29Z
blocked_by:
    - skull2-p22b
---

Sort the repo list by star count. Split out from the alphabet/last-updated
sorting bean because stars requires data-layer work the other two do not.

## Why its own bean

There is no `Stars` field on the `provider.Repo` model and no provider fetches
it. Sorting by stars therefore touches the whole stack, unlike name/updated
which already exist on the model.

## Scope

- Add `Stars int` to `provider.Repo` (`internal/provider/provider.go`).
- Populate it in all three clients:
  - `github.go`   -> `stargazers_count`
  - `codeberg.go` -> `stars_count`  (Forgejo/Gitea)
  - `gitlab.go`   -> `star_count`
- Cache is plain JSON: field is additive/backward-compatible (old caches read
  0 until refreshed). Add tests per provider.
- Add stars as a sort key in the TUI once the field exists (builds on the
  alphabet/last-updated sorting bean).

## Coordination

- The `Stars` model field may also be wanted by the repo-details view
  (add-repo-details). Whichever change lands the field first, the other reuses
  it — do not add it twice.
- Depends on / builds on the alphabet+last-updated sorting bean for the TUI
  sort plumbing.
