## Why

The provider abstraction and config exist but list nothing yet. Skull2 needs
real clients that fetch repositories from GitHub, Codeberg/Forgejo and GitLab,
and a JSON cache that persists that metadata as the single source of truth for
the TUI and sync.

## What Changes

- Implement the GitHub provider client (REST, pagination, `gh` CLI token or
  `SKULL2_GITHUB_TOKEN`, GHE `api_url` override).
- Implement the Codeberg/Forgejo/Gitea provider client (REST, pagination,
  `SKULL2_CODEBERG_TOKEN`, configurable `api_url`).
- Implement the GitLab provider client (v4 API, groups/subgroups + owned
  projects, `glab` token or `SKULL2_GITLAB_TOKEN`, self-hosted `api_url`).
- Register all three in the provider registry.
- Implement a per-provider JSON cache at `~/.cache/skull2/<provider>.json`
  (fetched_at + repos, atomic write) and a `skull2 refresh` command.

## Capabilities

### New Capabilities
- `provider-clients`: concrete GitHub, Codeberg and GitLab clients that list
  repositories via their HTTP APIs with pagination and auth.
- `repo-cache`: persistence and refresh of per-provider repository metadata.

### Modified Capabilities
<!-- none: provider-abstraction interface is consumed, not changed -->

## Impact

- `internal/provider` (github.go, codeberg.go, gitlab.go + registration).
- `internal/cache` (new implementation).
- `cmd/skull2` (new `refresh` subcommand).
- No new third-party dependencies expected (net/http + encoding/json).
