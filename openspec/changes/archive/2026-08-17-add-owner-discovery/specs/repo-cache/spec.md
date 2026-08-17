## ADDED Requirements

### Requirement: Owner index with per-owner fetch state

The cache SHALL record the set of known owners for a provider and, for each,
whether its repositories have been fetched, so consumers can distinguish
"discovered" from "fetched".

#### Scenario: Discovered-but-unfetched owner is representable

- **WHEN** an owner has been discovered but its repositories have not been
  fetched
- **THEN** the cache records the owner with an empty/unset fetch timestamp and no
  repositories for it

#### Scenario: Fetching an owner records its repos and timestamp

- **WHEN** an owner's repositories are fetched and cached
- **THEN** the cache stores those repositories and marks that owner as fetched
  with a timestamp, without discarding other owners' data

#### Scenario: Legacy flat cache is still readable

- **WHEN** an older cache file without an owner index is loaded
- **THEN** it is read successfully, treating every owner present in its
  repositories as already fetched
