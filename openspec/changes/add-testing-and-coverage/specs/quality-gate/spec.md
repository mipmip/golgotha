## ADDED Requirements

### Requirement: End-to-end coverage of the core flows

The system SHALL include end-to-end tests that prove the primary flows against
mocked provider APIs and temporary directories, without network access.

#### Scenario: Refresh, browse, clone

- **WHEN** the e2e test refreshes from a mocked provider and clones a selected
  repository
- **THEN** the repository is cloned to the path produced by the clone-path
  template

#### Scenario: Sync clone-missing then fast-forward

- **WHEN** the e2e test runs sync twice — first with no local clone, then after
  an upstream commit
- **THEN** the first run clones the repo and the second fast-forward-updates it

### Requirement: Coverage threshold enforced

The system SHALL measure test coverage and fail the build when it is below
target: overall at least 70% and core-logic packages at least 80%.

#### Scenario: Passing coverage

- **WHEN** coverage meets or exceeds the thresholds
- **THEN** `nix flake check` succeeds

#### Scenario: Failing coverage

- **WHEN** coverage falls below a threshold
- **THEN** the coverage check fails, failing `nix flake check`
