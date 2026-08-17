## ADDED Requirements

### Requirement: Repository visibility field

The `Repo` model SHALL include a `Visibility` string describing whether a
repository is public, private, or internal.

#### Scenario: Visibility is represented

- **WHEN** a repository is modeled
- **THEN** it carries a `Visibility` of `public`, `private`, or `internal`

#### Scenario: Unknown visibility normalizes

- **WHEN** a provider does not supply a recognizable visibility
- **THEN** it normalizes to `public`
