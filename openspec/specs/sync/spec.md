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

The system SHALL provide `hup sync [--provider NAME] [--no-refresh]` suitable
for cron.

#### Scenario: Refresh then sync

- **WHEN** the user runs `hup sync` without `--no-refresh`
- **THEN** the cache is refreshed before the engine runs

#### Scenario: Non-zero exit on failure

- **WHEN** any repository fails to clone or update
- **THEN** the command logs the failure line-by-line and exits non-zero

### Requirement: Eager discovery sweep on sync

When `all_owners` is enabled, `hup sync` and `hup refresh` SHALL discover
owners and fetch repositories for the entire resolved owner set, so the cache
and backups cover every org.

#### Scenario: Sync covers all discovered owners

- **WHEN** `hup sync` runs for a provider with `all_owners: true`
- **THEN** it discovers owners, fetches repositories for every owner in the
  resolved set (minus exclusions), updates the cache, and clones/pulls them

#### Scenario: Excluded owners are skipped

- **WHEN** an owner is listed in `exclude_owners`
- **THEN** sync neither fetches nor clones that owner's repositories

### Requirement: CLI fetch progress output

`hup sync` and `hup refresh` SHALL print progress derived from the fetch
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

### Requirement: Clone can emit progress

The clone engine SHALL offer a clone operation that emits progress events
(a percentage and a phase label) as the clone proceeds, so an interactive caller
can render a determinate progress bar. The existing non-progress clone used by
`hup sync` is unaffected.

#### Scenario: Progress is derived from git

- **WHEN** a progress-emitting clone runs
- **THEN** it invokes `git clone --progress` and parses git's progress output
  (e.g. "Receiving objects: N%") into percentage/phase events

#### Scenario: Completion and failure are reported

- **WHEN** a progress-emitting clone finishes or fails
- **THEN** the caller receives a terminal result (success with the resolved
  target, or an error) after the progress events

#### Scenario: Cancellation stops the clone

- **WHEN** the caller cancels the context of a progress-emitting clone
- **THEN** the underlying git process is stopped and the operation returns
  promptly without leaving a completed clone reported

### Requirement: Clone via jj when configured

When the resolved clone VCS for a repository is `jj`, the clone engine SHALL
clone it with `jj git clone --colocate` so a real top-level `.git` remains and
the existing sync operations (repo detection, dirty check, fast-forward pull)
continue to work unchanged. When the resolved VCS is `git`, it SHALL clone with
git as before.

#### Scenario: jj clone is colocated

- **WHEN** the resolved VCS for a repository is `jj`
- **THEN** the engine clones it with `jj git clone --colocate`, producing a
  working tree with a usable `.git`

#### Scenario: git clone unchanged

- **WHEN** the resolved VCS is `git`
- **THEN** the engine clones with git exactly as before

#### Scenario: jj missing is an actionable error

- **WHEN** a jj clone is requested but `jj` is not available on `PATH`
- **THEN** the clone fails with an error naming the missing `jj` tool

### Requirement: jj clone progress from a pseudo-terminal

The progress-emitting clone SHALL report determinate progress for jj clones.
Because `jj` emits its percentage only to a terminal and has no force-progress
flag, the engine SHALL run the jj clone attached to a pseudo-terminal, strip
ANSI control sequences, and parse the percentage into progress events, falling
back to indeterminate progress when a chunk cannot be parsed. The plain
(non-progress) clone used by `hup sync` needs no pseudo-terminal.

#### Scenario: jj progress is parsed from the pty

- **WHEN** a progress-emitting jj clone runs
- **THEN** the engine reads the pseudo-terminal output, strips ANSI, and emits
  percentage progress events

#### Scenario: Unparseable output falls back gracefully

- **WHEN** the jj clone output does not yield a parseable percentage
- **THEN** the clone still completes and reports its terminal result, with
  progress shown indeterminately rather than stuck

