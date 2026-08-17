## ADDED Requirements

### Requirement: Clients implement detail and README fetch

Each provider client SHALL implement repository detail and README fetching
against its API.

#### Scenario: GitHub details and README

- **WHEN** details/README are fetched for a GitHub repository
- **THEN** the client reads `stargazers_count`, `topics` and language, and the
  README via the repository README endpoint

#### Scenario: Codeberg details and README

- **WHEN** details/README are fetched for a Codeberg (Gitea) repository
- **THEN** the client reads stars/topics/language and the raw README

#### Scenario: GitLab details and README

- **WHEN** details/README are fetched for a GitLab project
- **THEN** the client reads `star_count`/topics/language and the README via the
  files API
