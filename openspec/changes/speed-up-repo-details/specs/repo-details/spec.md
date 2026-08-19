## MODIFIED Requirements

### Requirement: Lazy repository detail fetch

The system SHALL fetch a repository's extended details (stars, topics, language)
and its README on first open, and reuse them thereafter. The two requests SHALL
be issued concurrently, and details SHALL be shown even when the README request
fails.

#### Scenario: Details fetched on first open

- **WHEN** a repository's detail view is opened and no detail cache exists
- **THEN** the system fetches stars, topics, language and the README, then
  displays them

#### Scenario: The two requests run concurrently

- **WHEN** a cold detail fetch runs
- **THEN** the details request and the README request are issued concurrently
  (not one after the other)

#### Scenario: README failure still shows details

- **WHEN** the details request succeeds but the README request fails
- **THEN** the details are shown with an empty/placeholder README rather than
  failing the whole open

#### Scenario: Cached details reused

- **WHEN** a repository's detail view is opened and a detail cache exists
- **THEN** it is shown without re-fetching

#### Scenario: Manual refresh

- **WHEN** the user triggers refresh in the detail view
- **THEN** the repository's details and README are re-fetched and the detail
  cache updated

## ADDED Requirements

### Requirement: Detail prefetch warms the cache

The system SHALL be able to fetch a repository's details in the background to
warm the detail cache before the detail view is opened, using the same fetch and
cache path as an on-open fetch. A prefetch SHALL be skipped when the repository's
details are already cached.

#### Scenario: Prefetch populates the cache

- **WHEN** a repository is prefetched and no detail cache exists
- **THEN** its details are fetched and written to the detail cache, so a
  subsequent open is served from cache without a network fetch

#### Scenario: Prefetch skips already-cached repos

- **WHEN** a repository whose details are already cached is prefetched
- **THEN** no network fetch is performed
