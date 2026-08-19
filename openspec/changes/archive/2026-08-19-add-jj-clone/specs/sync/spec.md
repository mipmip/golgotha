## ADDED Requirements

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
