## Context

The provider abstraction (interface, Repo model, auth resolver, registry) and
config exist. This change adds the concrete HTTP clients and the JSON cache.

## Goals / Non-Goals

**Goals:**
- Three working, paginated, authenticated provider clients mapped to `Repo`.
- A robust per-provider JSON cache and `skull2 refresh`.
- Hermetic tests (mocked HTTP, temp cache dir).

**Non-Goals:**
- Sync/clone behavior (milestone 03) and the TUI (milestone 04).

## Decisions

- **HTTP**: standard library `net/http` + `encoding/json`; no SDKs, to avoid
  heavy dependencies for a PoC. Each client takes an injectable `*http.Client`
  and base URL so tests point at `httptest.Server`.
- **Pagination**: GitHub/Gitea via `Link` headers or page counting; GitLab via
  `X-Next-Page`/`page` params. Encapsulate per client.
- **Mapping**: each client maps its API shape to the shared `Repo`; central
  archived/fork filtering from the abstraction is reused.
- **GitLab owners**: treat a configured owner as a group; list group projects
  with `include_subgroups=true`, plus the authenticated user's owned projects.
- **Cache format**: `{ "fetched_at": RFC3339, "repos": [...] }`; write to a
  temp file in the same dir then `os.Rename` for atomicity; create
  `~/.cache/skull2` (honor `$XDG_CACHE_HOME`).

## Risks / Trade-offs

- [API shape drift across providers] → keep mapping small and covered by tests
  against recorded fixture JSON.
- [Rate limits with PAT] → out of scope for the PoC beyond honoring auth; note
  for later.
