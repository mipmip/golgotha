## ADDED Requirements

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
