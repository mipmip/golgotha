## ADDED Requirements

### Requirement: goreleaser builds multi-platform binaries

goreleaser SHALL build binaries for linux/amd64, linux/arm64, darwin/amd64 and
darwin/arm64, injecting the version from the git tag via ldflags, with checksums.

#### Scenario: goreleaser config valid

- **WHEN** `goreleaser check` is run
- **THEN** it passes with no errors

#### Scenario: Snapshot build succeeds

- **WHEN** `goreleaser build --snapshot --clean` is run
- **THEN** binaries are produced for all four platform targets

### Requirement: GitHub Actions triggers on version tags

A GitHub Actions workflow SHALL trigger on `v*` tag pushes and run goreleaser to
create a GitHub release with binaries and checksums.

#### Scenario: Workflow triggers on tag

- **WHEN** a tag matching `v*` is pushed
- **THEN** the release workflow runs goreleaser to publish the release

### Requirement: Gated release script

`scripts/release.sh` SHALL run safety checks before making changes, including a
clean working tree, being on `main`, an `Unreleased` CHANGELOG section, the
target tag not already existing, and a passing `nix flake check`.

#### Scenario: Dirty tree aborts

- **WHEN** `scripts/release.sh` runs with uncommitted changes
- **THEN** it exits with an error before modifying anything

#### Scenario: Failing gate aborts

- **WHEN** `nix flake check` fails during the safety checks
- **THEN** the release aborts before any commit or tag

### Requirement: Release script performs the release

The release script SHALL prompt for a version bump (major/minor/patch via gum),
update `VERSION`, promote the CHANGELOG `Unreleased` section to the new version
and date, update the flake `vendorHash`, then commit, tag and push.

#### Scenario: Release runs on a clean main

- **WHEN** the script runs on a clean `main` and all checks pass
- **THEN** it updates VERSION, CHANGELOG and flake vendorHash, creates a commit,
  creates and pushes a `vX.Y.Z` git tag

### Requirement: Release script updates the nix vendorHash

The script SHALL compute the correct `vendorHash` by building with a fake hash
and parsing the expected value; if nix is not installed it SHALL warn and skip.

#### Scenario: Nix available

- **WHEN** the script runs and nix is installed
- **THEN** `flake.nix` vendorHash is updated to the correct value

#### Scenario: Nix not available

- **WHEN** the script runs and nix is not installed
- **THEN** a warning is printed and the vendorHash update is skipped

### Requirement: Maintainer documentation

A `RELEASING.md` SHALL document prerequisites, the pre-release checklist, how to
run the release, how to verify it, and troubleshooting.

#### Scenario: Documentation exists

- **WHEN** `RELEASING.md` is read
- **THEN** it contains the pre-release checklist, release steps and
  troubleshooting

### Requirement: Release tooling in the dev shell

The flake devShell SHALL include goreleaser and gum.

#### Scenario: Dev shell has release tools

- **WHEN** `nix develop` is entered
- **THEN** goreleaser and gum are available
