## Context

TUI navigation is a strict tree: `levelProviders → levelOwners → levelRepos →
levelDetail`, with `selProvider`/`selOwner` tracking the drill-down. Repos live
in `reposByProvider map[string][]repoItem` (loaded from the per-provider cache),
each `repoItem` carrying its own `Provider` and resolved clone `Target`. Fetching
is **lazy**: `fetchedOwners` records which owners have been fetched;
`ownersByProvider` is the owner index (incl. discovered-but-unfetched owners).

`visibleRepos()` already has three branches; its third (`default`, when
`selProvider == nil`) iterates every provider and returns all their repos — but
it is unreachable in normal navigation. The combined view makes it reachable.

## Goals / Non-Goals

**Goals:**

- One flat, cross-provider list of all cached repos, reachable from the top.
- Reuse existing row behavior (filter/facets/sort/clone/detail) unchanged.
- Be honest about cache completeness; offer a full refresh to complete it.
- Stay independent of the in-flight config-schema changes.

**Non-Goals:**

- Streaming/background fetch (refresh is an explicit, foreground action).
- Columns (`skull2-n3i2`), star sort (`skull2-2h8p`), paging (`skull2-4pob`).
- The modes framework (`skull2-wzbf`).
- Self-account tint (needs `username` from `fix-self-owner-resolution`; follow-up).

## Decisions

### Decision: Synthetic "All repositories" entry, reusing `levelRepos`

Render a synthetic first row at `levelProviders`. Selecting it sets a `flatAll`
flag and enters `levelRepos` with `selProvider == nil` / `selOwner == ""`, so
`visibleRepos()` takes its existing all-providers branch. No new `level` value is
introduced.

- **Why:** matches "a view, not a new level"; reuses all existing repos-level
  rendering and key handling with the smallest surface change.
- **Alternative — a dedicated `levelAll`**: duplicates repos-level logic; more
  code, more places to keep in sync. Rejected.
- **Note:** code paths that assume `selProvider != nil` at `levelRepos` must
  branch on `flatAll` (rendering the row prefix, computing the badge, and routing
  refresh). Per-item actions already use `repoItem.Provider`, so clone/detail are
  unaffected.

### Decision: Show cached, badge completeness

The list is the flattened cache; completeness is `sum(fetchedOwners==true)` over
`sum(len(ownersByProvider))`. Display `N repos · X/Y owners loaded`. When the
owner index is empty (legacy cache without an index), fall back to `N repos`
without the fraction.

- **Why:** instant and honest; no new fetch machinery for the common path.

### Decision: Full refresh-all as an explicit foreground action

A refresh key in the flat view re-fetches **every** owner across **every**
provider, reusing the existing eager per-owner fetch, aggregating progress, and
updating both `reposByProvider` and the cache. It is cancellable; cancelling
leaves the prior cache intact.

- **Why:** the user explicitly opted for a full refresh (freshness) over a
  gap-fill; a portfolio overview is the use case where fetching everything is
  worth the burst, and making it explicit keeps the lazy default intact
  elsewhere.
- **Alternative — gap-fill only (fetch just unfetched owners)**: cheaper but
  never refreshes stale data. Could be added later as a lighter companion; not
  required now.

### Decision: Provider-prefixed rows

In `flatAll`, prefix each row with the provider short code before `owner/name`
(e.g. `gh  mipmip/foo`), keeping columns aligned. Fuzzy matching may include the
provider short so users can filter by provider token.

- **Why:** a bare `owner/name` is ambiguous across providers; the short code is
  the minimal disambiguator (full columns are `skull2-n3i2`).

## Risks / Trade-offs

- **[Full refresh burst]** → many owner fetches at once. Mitigation: it is
  explicit and cancellable, reuses the bounded-parallel per-owner fetch, and
  shows progress; the lazy per-owner default is unchanged everywhere else.
- **[`selProvider == nil` assumptions at `levelRepos`]** → latent nil-deref risk
  in paths that assumed a selected provider. Mitigation: audit those paths and
  branch on `flatAll`; add tests entering the flat view directly.
- **[Large flat list]** → hundreds of rows. Mitigation: existing windowing in
  `viewport.go` already handles large lists; paging keys are `skull2-4pob`.
- **[Ordering looks arbitrary]** → default is provider-then-cache order.
  Mitigation: acceptable for this scope; better default/interactive sort is
  `skull2-2h8p`.

## Migration Plan

1. Add the `flatAll` flag and the synthetic entry; route `visibleRepos()` and
   rendering.
2. Add the completeness badge and the full refresh-all command.
3. Ship without tint. Once `fix-self-owner-resolution` lands, add self-account
   tint in the flat view as a follow-up (owner == provider `username`).

## Open Questions

- Should a lighter "load only unfetched owners" companion to full refresh be
  added later? Deferred; full refresh covers the stated need.
