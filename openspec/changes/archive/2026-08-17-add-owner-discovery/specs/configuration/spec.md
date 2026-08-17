## ADDED Requirements

### Requirement: Owner discovery configuration

The system SHALL support a per-provider `all_owners` boolean and an
`exclude_owners` list, and SHALL resolve the effective owner set from them.

#### Scenario: all_owners resolves to self plus discovered orgs plus explicit owners

- **WHEN** a provider sets `all_owners: true`
- **THEN** the effective owner set is the union of the user's own account, every
  discovered organization, and any explicit `owners`, minus `exclude_owners`

#### Scenario: exclude_owners removes owners case-insensitively

- **WHEN** an owner name appears in `exclude_owners` (any case)
- **THEN** that owner is removed from the effective set, even the user's own
  account

#### Scenario: Default is unchanged behavior

- **WHEN** `all_owners` is unset or false
- **THEN** owner resolution uses the explicit `owners` list exactly as before
  (empty means the authenticated user's own repos)
