## MODIFIED Requirements

### Requirement: Load configuration from the default location

The system SHALL load configuration from `~/.config/golgotha/config.yaml`,
parsing it into typed global and per-provider settings.

#### Scenario: Valid config file is loaded

- **WHEN** `~/.config/golgotha/config.yaml` exists and is valid YAML matching
  the schema
- **THEN** the system returns a typed configuration containing `base_dir`, the
  clone-path template, and the list of providers with their settings

#### Scenario: Config file is missing

- **WHEN** no config file exists at the resolved path
- **THEN** the system returns an actionable error naming the expected path

#### Scenario: Config file is malformed

- **WHEN** the config file contains invalid YAML or unknown required fields
- **THEN** the system returns an error identifying the problem without panicking

### Requirement: Config CLI subcommands

The system SHALL expose the configuration through `gol config` subcommands.

#### Scenario: Print the config path

- **WHEN** the user runs `gol config path`
- **THEN** the system prints the resolved config file path and exits zero

#### Scenario: Validate the config

- **WHEN** the user runs `gol config check` with a valid config
- **THEN** the system prints a success summary and exits zero

#### Scenario: Report an invalid config

- **WHEN** the user runs `gol config check` with an invalid config
- **THEN** the system prints the validation error and exits non-zero
