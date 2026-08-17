# provider-clients Specification

## Purpose
TBD - created by archiving change add-provider-clients-and-cache. Update Purpose after archive.
## Requirements
### Requirement: GitHub client lists repositories

The system SHALL list repositories from GitHub for the configured user and
organizations via the REST API, following pagination.

#### Scenario: Paginated listing

- **WHEN** a configured owner has more repositories than one API page
- **THEN** the client follows pagination and returns all repositories mapped to
  the `Repo` model

#### Scenario: Auth and base URL

- **WHEN** a `gh` token or `SKULL2_GITHUB_TOKEN` is available and `api_url` is
  set for GitHub Enterprise
- **THEN** requests are authenticated and sent to the configured base URL

### Requirement: Codeberg client lists repositories

The system SHALL list repositories from a Forgejo/Gitea instance (Codeberg) via
its REST API, following pagination, authenticated by `SKULL2_CODEBERG_TOKEN`.

#### Scenario: Paginated listing

- **WHEN** listing repositories for the configured owners on Codeberg
- **THEN** the client follows pagination and returns all repositories mapped to
  the `Repo` model

### Requirement: GitLab client lists repositories

The system SHALL list projects from GitLab v4 across the configured groups and
subgroups plus owned projects, following pagination, authenticated by a `glab`
token or `SKULL2_GITLAB_TOKEN`.

#### Scenario: Groups and subgroups

- **WHEN** a configured owner is a group with subgroups
- **THEN** projects from the group and its subgroups are returned with the
  namespace mapped to the `Repo` owner

### Requirement: Clients honor repository filters

The system SHALL apply the `include_archived` and `include_forks` settings for
every provider client.

#### Scenario: Archived and forks excluded

- **WHEN** a provider sets `include_archived=false` and `include_forks=false`
- **THEN** archived and fork repositories are omitted from the listing

### Requirement: Owner discovery per provider

Each provider client SHALL discover the organizations the authenticated user
belongs to, for use when `all_owners` is enabled.

#### Scenario: GitHub and Codeberg discover member organizations

- **WHEN** discovery runs for a GitHub or Codeberg provider
- **THEN** the client lists the authenticated user's organizations (e.g.
  `/user/orgs`), following pagination

#### Scenario: GitLab discovers member groups

- **WHEN** discovery runs for a GitLab provider
- **THEN** the client lists the groups the user is a member of, following
  pagination

#### Scenario: Low discovery is surfaced

- **WHEN** `all_owners` is enabled but discovery returns no organizations
- **THEN** the system emits a warning (likely a missing token scope) rather than
  silently proceeding as if the user has no orgs

### Requirement: Page-aware, progress-emitting fetch

Each provider client SHALL fetch an owner's repositories page by page, emitting a
progress event per page, and determine the total page count when the API exposes
it.

#### Scenario: Total is derived from the first page

- **WHEN** the first page of an owner's repositories is fetched
- **THEN** the client derives the total page count from the provider's paging
  metadata (GitHub `Link rel="last"`, GitLab `X-Total-Pages`, Gitea
  `X-Total-Count`) when available

#### Scenario: Remaining pages fetched bounded-parallel

- **WHEN** the total page count is known and greater than one
- **THEN** the client fetches pages 2..N with at most the fixed worker cap in
  flight and merges the results, deduped by owner/name

#### Scenario: Fallback when total is unknown

- **WHEN** the provider does not expose a page total
- **THEN** the client falls back to sequential pagination, still emitting a page
  event per page

### Requirement: Non-progress fetch remains available

The system SHALL keep a simple fetch entry point that returns all repositories
without requiring an event consumer.

#### Scenario: Simple ListRepos still works

- **WHEN** a caller invokes the plain list operation
- **THEN** it returns the full repository slice as before (implemented over the
  event-emitting fetch)

### Requirement: Clients populate visibility

Each provider client SHALL populate `Repo.Visibility` from its API.

#### Scenario: GitHub and Codeberg map the private flag

- **WHEN** a GitHub or Codeberg repository is mapped
- **THEN** `Visibility` is `private` when the API marks it private, otherwise
  `public`

#### Scenario: GitLab maps its visibility value

- **WHEN** a GitLab project is mapped
- **THEN** `Visibility` reflects the project's `visibility`
  (`public`/`private`/`internal`)

