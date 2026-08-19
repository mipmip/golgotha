## ADDED Requirements

### Requirement: Provider username identifies the self account

Each provider configuration SHALL include a required `username` field naming the
authenticated user's own account on that provider. The system SHALL treat that
username as the identity of the self account for owner resolution, fetching and
display. There is no separate sentinel value for "the user's own account"; the
self account is an ordinary owner named by `username`.

#### Scenario: username is required

- **WHEN** a provider omits `username` (or sets it to an empty value)
- **THEN** validation fails naming the provider and the missing `username` field

#### Scenario: Self account participates like any owner

- **WHEN** owner resolution includes the self account
- **THEN** it appears under the configured `username`, indistinguishable in
  structure from an organization owner

## MODIFIED Requirements

### Requirement: Validate configuration

The system SHALL validate the configuration and reject invalid input with a
clear, actionable message.

#### Scenario: At least one provider is required

- **WHEN** the config defines no providers
- **THEN** validation fails with an error stating that at least one provider is
  required

#### Scenario: Provider names must be unique

- **WHEN** two providers share the same `name`
- **THEN** validation fails identifying the duplicated name

#### Scenario: Provider type must be known

- **WHEN** a provider has a `type` other than `github`, `codeberg` or `gitlab`
- **THEN** validation fails identifying the offending provider and type

#### Scenario: Required provider fields are present

- **WHEN** a provider is missing a required field (`name`, `type`, `short`, or
  `username`)
- **THEN** validation fails naming the missing field and the provider

### Requirement: Owner discovery configuration

The system SHALL support a per-provider `all_owners` boolean and an
`exclude_owners` list, and SHALL resolve the effective owner set from them. The
self account participates in resolution under the configured `username`.

#### Scenario: all_owners resolves to self plus discovered orgs plus explicit owners

- **WHEN** a provider sets `all_owners: true`
- **THEN** the effective owner set is the union of the configured `username`,
  every discovered organization, and any explicit `owners`, minus
  `exclude_owners`

#### Scenario: exclude_owners removes owners case-insensitively

- **WHEN** an owner name appears in `exclude_owners` (any case), including the
  configured `username`
- **THEN** that owner is removed from the effective set, even the user's own
  account

#### Scenario: Default is unchanged behavior

- **WHEN** `all_owners` is unset or false
- **THEN** owner resolution uses the explicit `owners` list, and an empty
  `owners` list resolves to exactly the configured `username` (the user's own
  repositories)
