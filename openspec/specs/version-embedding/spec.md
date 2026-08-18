# version-embedding Specification

## Purpose
TBD - created by archiving change add-release-process. Update Purpose after archive.
## Requirements
### Requirement: VERSION file as single source of truth

A `VERSION` file at the project root SHALL contain the version number (e.g.
`0.1.0`), with no `v` prefix.

#### Scenario: VERSION file exists

- **WHEN** the project root is inspected
- **THEN** a `VERSION` file exists containing a valid semver string

### Requirement: Binary embeds the version

The `hup version` command SHALL print the version embedded from the `VERSION`
file via `go:embed`, unless overridden at build time.

#### Scenario: Dev build reads VERSION

- **WHEN** `go run ./cmd/hup version` is executed
- **THEN** it prints the version from the `VERSION` file

#### Scenario: ldflags override wins

- **WHEN** the binary is built with `-ldflags "-X main.version=v1.2.3"`
- **THEN** `hup version` prints `v1.2.3` (ldflags take precedence for
  goreleaser)

### Requirement: Flake reads the version

The `flake.nix` SHALL read the version from the `VERSION` file via
`builtins.readFile`, not a hardcoded string.

#### Scenario: Nix build uses VERSION

- **WHEN** `nix build` runs
- **THEN** the built binary reports the version from the `VERSION` file

