## MODIFIED Requirements

### Requirement: Load configuration from the default location

The system SHALL load configuration from `~/.config/huphop/config.yaml`,
parsing it into typed global and per-provider settings.

#### Scenario: Valid config file is loaded

- **WHEN** `~/.config/huphop/config.yaml` exists and is valid YAML matching
  the schema
- **THEN** the system returns a typed configuration containing `base_dir`, the
  clone-path template, and the list of providers with their settings

#### Scenario: Config file is missing

- **WHEN** no config file exists at the resolved path
- **THEN** the system returns an actionable error naming the expected path

#### Scenario: Config file is malformed

- **WHEN** the config file contains invalid YAML or unknown required fields
- **THEN** the system returns an error identifying the problem without panicking
