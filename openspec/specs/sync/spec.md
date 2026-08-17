# sync Specification

## Purpose
TBD - created by archiving change add-cli-sync. Update Purpose after archive.
## Requirements
### Requirement: Clone missing repositories

The system SHALL clone a repository to its templated target path when that path
does not yet contain a git repository, using the provider's clone protocol.

#### Scenario: Missing repo is cloned

- **WHEN** a cached repository has no local clone at its resolved target path
- **THEN** the system clones it there using the configured clone protocol

### Requirement: Fast-forward pull existing repositories

The system SHALL update existing clones with a fetch and fast-forward-only pull
on the default branch, never forcing.

#### Scenario: Clean repo is fast-forwarded

- **WHEN** a local clone exists and is clean and behind its remote
- **THEN** the system fetches and fast-forward-pulls the default branch

#### Scenario: Dirty repo is skipped

- **WHEN** a local clone has uncommitted changes
- **THEN** the system skips it and records a warning instead of modifying it

### Requirement: Sync summary

The system SHALL report a per-provider summary of the run.

#### Scenario: Summary counts outcomes

- **WHEN** a sync run completes
- **THEN** it reports counts of cloned, updated, skipped and failed repositories
  per provider

### Requirement: Sync command

The system SHALL provide `skull2 sync [--provider NAME] [--no-refresh]` suitable
for cron.

#### Scenario: Refresh then sync

- **WHEN** the user runs `skull2 sync` without `--no-refresh`
- **THEN** the cache is refreshed before the engine runs

#### Scenario: Non-zero exit on failure

- **WHEN** any repository fails to clone or update
- **THEN** the command logs the failure line-by-line and exits non-zero

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

### Requirement: CLI fetch progress output

`skull2 sync` and `skull2 refresh` SHALL print progress derived from the fetch
event stream.

#### Scenario: Per-owner progress lines

- **WHEN** the CLI fetches repositories for owners
- **THEN** it prints line-oriented progress per owner (e.g. start, completion
  with a repo count, and any warnings), remaining cron-friendly

### Requirement: Bounded-parallel owner fetch on the CLI

The CLI SHALL fetch multiple owners concurrently within the fixed worker cap.

#### Scenario: Owners fetched concurrently

- **WHEN** many owners must be fetched
- **THEN** at most the fixed worker cap are fetched at once, and the printed
  progress still attributes each line to its owner

