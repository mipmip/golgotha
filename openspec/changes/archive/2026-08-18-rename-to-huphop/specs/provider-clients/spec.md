## MODIFIED Requirements

### Requirement: GitHub client lists repositories

The system SHALL list repositories from GitHub for the configured user and
organizations via the REST API, following pagination.

#### Scenario: Paginated listing

- **WHEN** a configured owner has more repositories than one API page
- **THEN** the client follows pagination and returns all repositories mapped to
  the `Repo` model

#### Scenario: Auth and base URL

- **WHEN** a `gh` token or `HUPHOP_GITHUB_TOKEN` is available and `api_url` is
  set for GitHub Enterprise
- **THEN** requests are authenticated and sent to the configured base URL

### Requirement: Codeberg client lists repositories

The system SHALL list repositories from a Forgejo/Gitea instance (Codeberg) via
its REST API, following pagination, authenticated by `HUPHOP_CODEBERG_TOKEN`.

#### Scenario: Paginated listing

- **WHEN** listing repositories for the configured owners on Codeberg
- **THEN** the client follows pagination and returns all repositories mapped to
  the `Repo` model

### Requirement: GitLab client lists repositories

The system SHALL list projects from GitLab v4 across the configured groups and
subgroups plus owned projects, following pagination, authenticated by a `glab`
token or `HUPHOP_GITLAB_TOKEN`.

#### Scenario: Groups and subgroups

- **WHEN** a configured owner is a group with subgroups
- **THEN** projects from the group and its subgroups are returned with the
  namespace mapped to the `Repo` owner
