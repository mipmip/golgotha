---
# skull2-39es
title: Combined cross-provider flat repo list
status: completed
type: feature
priority: normal
created_at: 2026-08-19T12:52:56Z
updated_at: 2026-08-19T19:21:16Z
parent: skull2-ok4c
---

A single flat list of repositories across **all providers and all owners/orgs**
in one view (pim's wish, split off from the self-owner bug fix `skull2-qp5y`).

## Motivation

Today the TUI navigation is strictly hierarchical: provider → owner → repos
(`levelProviders` → `levelOwners` → `levelRepos`). There is no way to see a
combined list of every repo across providers and organizations at once. pim
wants exactly that combined view — the shape the old buggy "(your account)"
view accidentally resembled.

## Notes / open design questions

- This is a **new navigation mode / view**, not a new level in the existing
  tree. Likely relates to the modes concept in `skull2-wzbf` (management vs
  multiplex modes).
- `visibleRepos()` already has a latent "all providers" branch
  (`selProvider == nil`) that iterates every provider — currently unreachable
  in normal nav. A combined view could build on that.
- Scope: cross-provider AND cross-owner, flat, with the existing fuzzy filter
  and facet filters applying across the whole set.
- Decide how repos are labeled/disambiguated (provider short + owner prefix)
  and how the self account is tinted here too (see `skull2-qp5y`).

## Related

- Split off from `skull2-qp5y` (self-owner resolution / tint).
- See the cross-reference note in `skull2-wzbf` (tui-modes / multiplexer).


## Decisions (explore 2026-08-19)

- **Entry:** synthetic "All repositories" row pinned atop the provider list;
  selecting it enters the existing `levelRepos` with a `flatAll` flag
  (`selProvider=nil`) so `visibleRepos()` takes the latent all-providers branch.
  No new nav level type.
- **Data completeness:** show whatever is cached (flatten `reposByProvider`)
  instantly, with a completeness badge (`N repos · X/Y owners loaded`) from
  `ownersByProvider` vs `fetchedOwners`.
- **Refresh:** a **full** refresh-all across every provider×owner must be
  possible (re-fetch, freshness), reusing the existing eager-fetch machinery.
- **Rows:** minimal `provider-short owner/name` prefix for disambiguation;
  existing sort/filter/facets/clone/detail reused unchanged.
- **Self-tint:** deferred to a follow-up after `skull2-qp5y` (needs the resolved
  `username`); the combined view ships without tint so it stays config-independent.
- **Out of scope:** columns (skull2-n3i2), star sort (skull2-2h8p), paging
  (skull2-4pob), modes framework (skull2-wzbf).
- Change: `add-combined-repo-view` (tui-only spec delta).



## Summary of Changes

Added the combined cross-provider flat repo view. A synthetic "All repositories"
entry (appended to the provider list) sets a `flatAll` scope and enters
levelRepos with selProvider=nil, so visibleRepos aggregates every cached repo
across providers/owners. Rows are provider-short prefixed; fuzzy match includes
the provider short. Header shows a completeness badge (N repos · X/Y owners
loaded, with a refresh hint). `r` in the flat view does a full refresh across
all providers (batched `refresher`, per-provider atomic). Esc returns to the
provider list. Self-tint deferred (rides qp5y, already shipped — could be a
follow-up). Coverage overall 81.4%. Shipped as 2026-08-19-add-combined-repo-view
(commit 275da606).
