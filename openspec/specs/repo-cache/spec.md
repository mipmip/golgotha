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

