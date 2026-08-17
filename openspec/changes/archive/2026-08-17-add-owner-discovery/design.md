## Context

Listing is driven by an explicit `owners` list. Empty-owners semantics already
differ per provider (GitHub `/user/repos?affiliation=owner` = own only; GitLab
`/projects?membership=true` = broad). Cache is one `<provider>.json` of
`{fetched_at, repos[]}`. The TUI derives the owner level from distinct owners in
that cache and navigates provider → owner → repos. This change adds automatic
owner discovery and a split eager/lazy fetch model over a shared cache.

## Goals / Non-Goals

**Goals:**
- One opt-in switch to include every org you belong to; a declaration to ignore
  specific ones.
- Fast TUI (instant owner list, fetch-per-owner on entry).
- Complete backups (sync still sweeps everything).
- Non-breaking config.

**Non-Goals:**
- Outside-collaborator repos (repos shared with you outside your orgs) — a
  separate affiliation axis, deferred.
- Changing existing explicit-`owners` behavior beyond unioning discovery in.

## Decisions

- **Config (approach B, non-breaking):**
  ```yaml
  all_owners: true            # discover every org you belong to + your own account
  exclude_owners: [noisy-org] # ignore by declaration (case-insensitive; may include self)
  owners: [extra]             # still works; unioned in
  ```
  Owner-set resolution when `all_owners`:
  `{ self } ∪ { discovered orgs } ∪ { explicit owners } − { exclude_owners }`
  (case-insensitive). When `all_owners` is false: today's behavior unchanged.
- **`all_owners` includes the user's own account** as one owner (so you get your
  personal repos too). Excludable via `exclude_owners`.
- **Member orgs only** for v1 (GitHub/Codeberg `/user/orgs`; GitLab groups you're
  a member of). Collaborator repos are out of scope.
- **Split by command over one cache:**
  ```
  discovery ── owner list (cheap) ──▶ cached owner index
       sync/refresh → EAGER: fetch repos for every resolved owner (backup + full cache)
       tui          → LAZY:  show owners now; on entering an owner with no fetched
                             repos, fetch + cache; already-fetched owners are instant
  ```
- **Cache shape (v2):** keep repos but add an owner index with per-owner state, e.g.
  ```json
  { "discovered_at": "...",
    "owners": [ { "name": "acme", "fetched_at": "..." },
                { "name": "mipmip", "fetched_at": null } ],
    "repos": [ ... ] }
  ```
  `fetched_at: null` ⇒ discovered but repos not yet fetched (TUI fetches on entry;
  sync fetches all). Backward-compatible read of the old flat shape (treat all
  present owners as fetched).
- **TUI:** owner level lists the cached owner index (not only owners already
  having repos). Entering an unfetched owner emits a fetch `tea.Cmd`, shows a
  loading line, then populates + caches. `r` re-fetches the current owner.
- **Discovery scope caveat:** with `all_owners` on, if discovery returns zero
  orgs, warn (likely a missing `read:org`-style scope) rather than silently
  backing up nothing new.

## Risks / Trade-offs

- [Silent under-discovery from missing token scope] → emit a warning on
  zero/low discovery; document required scopes.
- [Cache schema change] → version/tolerate the old flat shape on read.
- [Many orgs → many API calls on sync] → acceptable for cron; lazy TUI avoids it
  interactively.
- [Consistency across providers] → normalize so `all_owners` means the same
  thing everywhere even though GitLab's default is already broad.
