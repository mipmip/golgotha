## ADDED Requirements

### Requirement: Commit an owner only on complete fetch

The cache SHALL be updated for an owner only when that owner's fetch completes
successfully, so a cancelled or failed fetch never leaves a partial owner.

#### Scenario: Complete fetch commits the owner

- **WHEN** an owner's repositories are fully fetched
- **THEN** the cache stores them and marks the owner fetched with a timestamp

#### Scenario: Cancelled or failed fetch leaves the owner unfetched

- **WHEN** an owner's fetch is cancelled or any page fails
- **THEN** the cache is not modified for that owner; it remains unfetched and is
  re-fetched on the next attempt
