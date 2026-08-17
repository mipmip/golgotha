# repo-cache Specification

## Purpose
TBD - created by archiving change add-provider-clients-and-cache. Update Purpose after archive.
## Requirements
### Requirement: Persist per-provider repository cache

The system SHALL persist each provider's repositories to
`~/.cache/skull2/<provider>.json` with a fetch timestamp, writing atomically.

#### Scenario: Write and read round-trip

- **WHEN** a provider's repositories are cached and then loaded
- **THEN** the loaded data equals what was written, including `fetched_at` and
  every repository field

#### Scenario: Atomic write

- **WHEN** a cache file is written
- **THEN** it is written via a temporary file and rename so a concurrent reader
  never sees a partial file

#### Scenario: Missing cache

- **WHEN** no cache file exists for a provider
- **THEN** loading reports an empty/absent cache without error to callers that
  tolerate it

### Requirement: Refresh command

The system SHALL provide `skull2 refresh` to re-fetch repositories from the
providers and update the cache.

#### Scenario: Refresh all providers

- **WHEN** the user runs `skull2 refresh`
- **THEN** every configured provider is listed and its cache file is updated

#### Scenario: Refresh a single provider

- **WHEN** the user runs `skull2 refresh --provider NAME`
- **THEN** only that provider is refreshed, and an unknown name exits non-zero

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

