## ADDED Requirements

### Requirement: Eager discovery sweep on sync

When `all_owners` is enabled, `skull2 sync` and `skull2 refresh` SHALL discover
owners and fetch repositories for the entire resolved owner set, so the cache
and backups cover every org.

#### Scenario: Sync covers all discovered owners

- **WHEN** `skull2 sync` runs for a provider with `all_owners: true`
- **THEN** it discovers owners, fetches repositories for every owner in the
  resolved set (minus exclusions), updates the cache, and clones/pulls them

#### Scenario: Excluded owners are skipped

- **WHEN** an owner is listed in `exclude_owners`
- **THEN** sync neither fetches nor clones that owner's repositories
