## Context

Repo-level Enter clones today (overloaded with `c`); the `Repo` model lacks
stars/topics/language/README. This change adds a detail view reached by Enter,
with lazily-fetched details cached apart from the lean list cache.

## Goals / Non-Goals

**Goals:**
- Inspect a repo (metadata + rendered README) before cloning.
- Lazy, cached detail fetch that keeps the list cache lean.
- Consistent navigation (Enter drills; `c` clones).
- Graceful offline.

**Non-Goals:**
- Editing/creating repos; issues/PRs.
- Caching README inside the main `<provider>.json`.
- A TTL/auto-refresh for details (manual `r` only).

## Decisions

- **Enter remap:** repo-level Enter opens the detail view; `c` is the only clone
  key. Enter now means "drill deeper" at every level; Esc backs out.
- **Rendering: glamour** (in-process) instead of the bean's `bat`. Avoids
  embedding external ANSI and syncing terminal width; renders markdown→styled
  string, width-aware, in a `bubbles/viewport` (scrollable README). Cost: a new
  `github.com/charmbracelet/glamour` dependency and a flake vendorHash update.
- **Data tiers:** tier-1 from the existing cached `Repo` (free); tier-2
  (stars, topics, language) + tier-3 (README) fetched lazily on first open.
- **Separate detail cache:** `~/.cache/skull2/details/<provider>/<owner>__<repo>.json`
  = `{ fetched_at, stars, topics, language, readme_md }`. Store RAW markdown and
  render with glamour at view time. Never in the main list cache. `r` re-fetches
  the current repo's detail.
- **Provider methods:** add `RepoDetails(ctx, owner, name) (Details, error)` and
  `Readme(ctx, owner, name) (string, error)` (or one combined call). Map per
  provider: GitHub `stargazers_count`/`topics`/`/readme`; Gitea
  `stars_count`/topics/raw; GitLab `star_count`/topics/files.
- **Streaming/laziness:** the fetch is a Bubble Tea command returning a
  detailLoadedMsg; the view shows a loading line until it arrives (reuse the
  spinner from add-fetch-progress if landed). Update stays pure/testable via
  injected messages/fakes; no network in tests.
- **Graceful offline:** if the fetch errors and no detail cache exists, show
  tier-1 metadata plus a "README unavailable" note — no error screen. If a cache
  file exists, use it.

## Risks / Trade-offs

- [New dependency (glamour)] → vendorHash bump; justified by clean in-process
  rendering vs external-ANSI fragility.
- [README size] → separate per-repo files keep the list cache lean; raw markdown
  only.
- [Provider README endpoints differ] → small per-provider fetch; unavailable
  README degrades gracefully.
- [Client-file overlap with queued changes] → sequence last; this mostly ADDS
  methods rather than editing ListRepos.
