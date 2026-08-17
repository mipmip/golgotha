## ADDED Requirements

### Requirement: Load configuration from the default location

The system SHALL load configuration from `~/.config/skull2/config.yaml`,
parsing it into typed global and per-provider settings.

#### Scenario: Valid config file is loaded

- **WHEN** `~/.config/skull2/config.yaml` exists and is valid YAML matching the
  schema
- **THEN** the system returns a typed configuration containing `base_dir`, the
  clone-path template, and the list of providers with their settings

#### Scenario: Config file is missing

- **WHEN** no config file exists at the resolved path
- **THEN** the system returns an actionable error naming the expected path

#### Scenario: Config file is malformed

- **WHEN** the config file contains invalid YAML or unknown required fields
- **THEN** the system returns an error identifying the problem without panicking

### Requirement: Apply configuration defaults

The system SHALL apply documented defaults for values omitted from the config
file.

#### Scenario: Defaults are filled in

- **WHEN** the config omits `base_dir`, `clone_protocol`, `include_archived`,
  `include_forks`, or `clone_pattern_tpl`
- **THEN** the loaded configuration uses `base_dir=~`, `clone_protocol=ssh`,
  `include_archived=false`, `include_forks=true`, and the default
  `clone_pattern_tpl` of `{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}`

#### Scenario: base_dir tilde is expanded

- **WHEN** `base_dir` is `~` or begins with `~/`
- **THEN** the loaded configuration expands it to the user's home directory as
  an absolute path

### Requirement: Validate configuration

The system SHALL validate the configuration and reject invalid input with a
clear, actionable message.

#### Scenario: At least one provider is required

- **WHEN** the config defines no providers
- **THEN** validation fails with an error stating that at least one provider is
  required

#### Scenario: Provider names must be unique

- **WHEN** two providers share the same `name`
- **THEN** validation fails identifying the duplicated name

#### Scenario: Provider type must be known

- **WHEN** a provider has a `type` other than `github`, `codeberg` or `gitlab`
- **THEN** validation fails identifying the offending provider and type

#### Scenario: Required provider fields are present

- **WHEN** a provider is missing a required field (`name`, `type`, or `short`)
- **THEN** validation fails naming the missing field and the provider

### Requirement: Config CLI subcommands

The system SHALL expose the configuration through `skull2 config` subcommands.

#### Scenario: Print the config path

- **WHEN** the user runs `skull2 config path`
- **THEN** the system prints the resolved config file path and exits zero

#### Scenario: Validate the config

- **WHEN** the user runs `skull2 config check` with a valid config
- **THEN** the system prints a success summary and exits zero

#### Scenario: Report an invalid config

- **WHEN** the user runs `skull2 config check` with an invalid config
- **THEN** the system prints the validation error and exits non-zero
