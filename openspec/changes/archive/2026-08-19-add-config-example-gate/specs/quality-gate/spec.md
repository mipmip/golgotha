## ADDED Requirements

### Requirement: Documented example config stays in sync with the schema

The system SHALL verify the repository's `config.example.yaml` against the
config schema as part of `nix flake check`, failing the build when the example
is invalid or incomplete. The example is both a user-facing document and a
tested artifact.

#### Scenario: Example must be valid

- **WHEN** `config.example.yaml` is loaded and validated through the standard
  config loader
- **THEN** the check fails if loading or validation returns an error (including
  an unknown or misspelled key rejected by strict decoding)

#### Scenario: Example must document every schema field

- **WHEN** the check collects every configuration field name from the config
  schema types
- **THEN** it fails if any field name does not appear as a key in the example
  text, where a commented `# key:` line counts as documented

#### Scenario: Deliberate omissions are allow-listed

- **WHEN** a field is intentionally not documented in the example
- **THEN** it is recorded in an explicit allow-list so the completeness check
  passes, keeping omissions visible and intentional

#### Scenario: Gate runs without extra CI wiring

- **WHEN** `nix flake check` runs the existing test suite
- **THEN** the example gate runs as part of it and its failure fails the build
