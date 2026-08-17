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

