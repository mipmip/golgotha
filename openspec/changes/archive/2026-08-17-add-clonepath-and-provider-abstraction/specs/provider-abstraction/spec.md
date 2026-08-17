## ADDED Requirements

### Requirement: Repository domain model

The system SHALL define a provider-agnostic `Repo` model capturing the fields
needed by the cache, sync and TUI.

#### Scenario: Repo carries the required fields

- **WHEN** a provider returns a repository
- **THEN** it is represented with owner, name, description, ssh URL, https URL,
  web URL, default branch, archived flag, fork flag, and updated-at time

### Requirement: Provider interface

The system SHALL define a `Provider` interface that lists repositories for the
configured owners.

#### Scenario: Listing returns repositories

- **WHEN** a provider's list operation is invoked for a set of owners
- **THEN** it returns the repositories as `Repo` values or an error

#### Scenario: Filters are honored

- **WHEN** a provider is configured with `include_archived=false` or
  `include_forks=false`
- **THEN** archived or fork repositories are excluded from the result

### Requirement: Auth resolution

The system SHALL resolve a provider's credential in the order: configured CLI
token, then env-var PAT, then a clear error.

#### Scenario: CLI token is used when available

- **WHEN** a provider configures an auth `cli` whose token can be obtained
- **THEN** that token is used

#### Scenario: Env PAT fallback

- **WHEN** no CLI token is available but the configured env var is set
- **THEN** the PAT from the env var is used

#### Scenario: No credential available

- **WHEN** neither a CLI token nor the env PAT is available
- **THEN** the system returns an actionable error naming the provider and the
  expected env var

### Requirement: Provider registry

The system SHALL construct the correct provider client from a provider's
`type`.

#### Scenario: Known type resolves

- **WHEN** building a provider of type `github`, `codeberg` or `gitlab`
- **THEN** the corresponding client is returned

#### Scenario: Unknown type errors

- **WHEN** building a provider of an unknown type
- **THEN** the system returns an error identifying the type
