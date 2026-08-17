## ADDED Requirements

### Requirement: Owner level lists discovered owners

The TUI SHALL populate the owner level from the cached owner index, including
owners whose repositories have not yet been fetched.

#### Scenario: Discovered owners appear before their repos are fetched

- **WHEN** `all_owners` is enabled and owners have been discovered but not all
  fetched
- **THEN** the owner level lists every discovered owner, selectable regardless of
  whether its repositories are cached yet

### Requirement: Lazy per-owner repository fetch

The TUI SHALL fetch an owner's repositories the first time the user enters it and
cache the result; subsequent visits use the cache.

#### Scenario: Fetch on first entry

- **WHEN** the user enters an owner whose repositories have not been fetched
- **THEN** the TUI fetches them, shows a loading indicator while doing so, then
  displays and caches them

#### Scenario: Cached owner is instant

- **WHEN** the user enters an owner whose repositories are already cached
- **THEN** the repositories display immediately without re-fetching

#### Scenario: Refresh re-fetches the current owner

- **WHEN** the user triggers refresh while viewing an owner
- **THEN** that owner's repositories are re-fetched and the cache updated
