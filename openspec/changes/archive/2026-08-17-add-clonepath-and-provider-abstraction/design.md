## Context

Configuration is loaded (add-config-loading). These two capabilities complete
milestone 01 and are pure, dependency-free logic that milestone 02's provider
clients build on.

## Goals / Non-Goals

**Goals:**
- Deterministic clone-path rendering with the documented fields and safety.
- A small, testable provider abstraction and auth resolver with no network.

**Non-Goals:**
- Concrete GitHub/Codeberg/GitLab HTTP clients (milestone 02).
- The cache and sync (milestones 02/03).

## Decisions

- **text/template**: use the standard library `text/template` with a struct of
  the documented fields; no third-party templating. Missing keys error via
  `Option("missingkey=error")`.
- **Path safety**: render, expand `~`/base_dir, `filepath.Clean`, then verify
  the result is within `base_dir` using a prefix check on cleaned absolute
  paths; reject otherwise.
- **Provider interface**: keep it minimal — `ListRepos(ctx, owners) ([]Repo,
  error)` — so clients are easy to fake in tests. Filtering by
  archived/fork is applied centrally so every client behaves consistently.
- **Auth resolver**: a function taking a provider config and an environment
  lookup + a CLI-token getter (both injectable for tests), returning the token
  or an actionable error. CLI-token retrieval shells out to `gh`/`glab` and is
  behind an interface so unit tests avoid the network.
- **Registry**: a map from `ProviderType` to a constructor, populated by the
  concrete clients in milestone 02; unknown types error.

## Risks / Trade-offs

- [Registry has no real clients yet in M01] → M01 registers a fake type for
  tests; M02 registers the real github/codeberg/gitlab constructors.
- [CLI-token shelling is environment-dependent] → injected interface keeps unit
  tests hermetic; real behavior is covered by milestone 05 e2e.
