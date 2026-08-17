## ADDED Requirements

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
