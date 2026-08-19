## ADDED Requirements

### Requirement: Clone VCS selection

The system SHALL support choosing the version-control tool used to clone a
repository: a global `clone_vcs` (`git` or `jj`, default `git`), a per-provider
`clone_vcs` override, and per-repo `vcs_rules` (each a glob matched against
`owner/name` and a target `vcs`). The effective VCS for a repository SHALL be
resolved as: the first matching `vcs_rules` entry, else the provider's
`clone_vcs`, else the global `clone_vcs`, else `git`.

#### Scenario: Default is git

- **WHEN** no `clone_vcs` or `vcs_rules` are configured
- **THEN** repositories are cloned with git

#### Scenario: Per-repo rule wins over provider and global

- **WHEN** a repository's `owner/name` matches a provider `vcs_rules` entry
- **THEN** that rule's `vcs` is used, overriding the provider and global
  `clone_vcs`

#### Scenario: Provider override wins over global

- **WHEN** a provider sets `clone_vcs` and no `vcs_rules` entry matches the
  repository
- **THEN** the provider's `clone_vcs` is used, overriding the global default

#### Scenario: Invalid configuration is rejected

- **WHEN** a `clone_vcs`/`vcs_rules` `vcs` value is not `git` or `jj`, or a
  `match` glob is invalid
- **THEN** validation fails with an actionable message identifying the offending
  value
