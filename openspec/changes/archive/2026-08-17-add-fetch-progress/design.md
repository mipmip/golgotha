## Context

Provider clients paginate internally and return a finished `[]Repo`, so nothing
observes progress and nothing can cancel mid-fetch. The TUI fetches an owner
lazily on entry (`ownerFetchedMsg`) with a silent loading line; the CLI
`refresh`/`sync` print only a final per-provider count. This change adds a
shared progress-event model, bounded-parallel page fetching, and cancellation,
without corrupting the cache.

## Goals / Non-Goals

**Goals:**
- One progress model, two frontends (TUI + CLI), plus test observability.
- Faster fetch via bounded parallelism; visible textual progress.
- Safe cancellation with no partial/corrupt cache.

**Non-Goals:**
- Parallelizing discovery (`ListOwners`) — it is cheap.
- A configurable worker cap (fixed at 6 for now).
- Persisting partial owner results.

## Decisions

- **Event model (new `fetch-progress` capability):**
  ```
  FetchStarted{Provider, Owner}
  PageDone{Provider, Owner, Page, TotalPages, ReposSoFar}
  FetchDone{Provider, Owner, Count}
  FetchFailed{Provider, Owner, Err}
  Canceled{Provider, Owner}
  Warning{Provider, Owner, Msg}
  ```
  Emitted through an `emit func(Event)` (or a channel) passed into the fetch.
  Presentation is entirely in the frontends.
- **Page-parallel fetch (provider-clients):** fetch page 1 to learn the total
  (`Link rel="last"` for GitHub, `X-Total-Pages` for GitLab, `X-Total-Count`/
  page-size for Gitea), then fetch pages `2..N` with a bounded pool (cap 6),
  emitting `PageDone` as each returns. Results merged and deduped (existing
  owner/name de-dup). Providers keep a simple `ListRepos` wrapper for callers
  that don't want events.
- **Owner-parallel (CLI):** `sync`/`refresh` fetch owners with a bounded pool
  (cap 6); each owner's events are serialized into printed lines.
- **TUI (case c):** on entering an unfetched owner, run the page-aware fetch as a
  Bubble Tea command that streams events → `progressMsg`; render `bubbles/spinner`
  while the total is unknown, then a determinate `bubbles/progress` bar plus
  "fetching <owner> page i/n — N repos". `Esc` cancels the current fetch.
- **Cancellation + cache safety:**
  - `cache.Save` is already atomic (temp + rename) → no torn file.
  - Semantic guard (repo-cache): call `MarkOwnerFetched` only on a COMPLETE
    fetch. On cancel or any page failure, do not touch that owner — it stays
    `fetched_at = nil` and is re-fetched next entry.
  - The TUI shows partial results as transient (never cached as the owner's set).
- **Streaming in Bubble Tea:** a goroutine runs the fetch and sends events on a
  channel; a `tea.Cmd` waits on the channel and re-issues itself per event, so
  `Update` stays pure and is tested by feeding `progressMsg` values.

## Risks / Trade-offs

- [Page-parallel needs page 1 first] → sequential first page, then fan out;
  acceptable and works across all three providers.
- [Refactoring the client fetch is the bulk of the work] → keep the old
  `ListRepos` as a thin wrapper over the new event-emitting fetch to limit blast
  radius.
- [Cancel races on the channel] → cancel via context; drain/close the channel;
  the commit-only-on-complete rule makes a lost race harmless (owner just stays
  unfetched).
- [Provider total headers vary] → each client computes its own total; if a total
  is unavailable, fall back to sequential pagination with indeterminate progress.
