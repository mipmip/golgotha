# configuration Specification

## Purpose
TBD - created by archiving change add-config-loading. Update Purpose after archive.
## Requirements
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

- **WHEN** a provider is missing a required field (`name`, `type`, `short`, or
  `username`)
- **THEN** validation fails naming the missing field and the provider

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

### Requirement: Owner discovery configuration

The system SHALL support a per-provider `all_owners` boolean and an
`exclude_owners` list, and SHALL resolve the effective owner set from them. The
self account participates in resolution under the configured `username`.

#### Scenario: all_owners resolves to self plus discovered orgs plus explicit owners

- **WHEN** a provider sets `all_owners: true`
- **THEN** the effective owner set is the union of the configured `username`,
  every discovered organization, and any explicit `owners`, minus
  `exclude_owners`

#### Scenario: exclude_owners removes owners case-insensitively

- **WHEN** an owner name appears in `exclude_owners` (any case), including the
  configured `username`
- **THEN** that owner is removed from the effective set, even the user's own
  account

#### Scenario: Default is unchanged behavior

- **WHEN** `all_owners` is unset or false
- **THEN** owner resolution uses the explicit `owners` list, and an empty
  `owners` list resolves to exactly the configured `username` (the user's own
  repositories)

### Requirement: Provider username identifies the self account

Each provider configuration SHALL include a required `username` field naming the
authenticated user's own account on that provider. The system SHALL treat that
username as the identity of the self account for owner resolution, fetching and
display. There is no separate sentinel value for "the user's own account"; the
self account is an ordinary owner named by `username`.

#### Scenario: username is required

- **WHEN** a provider omits `username` (or sets it to an empty value)
- **THEN** validation fails naming the provider and the missing `username` field

#### Scenario: Self account participates like any owner

- **WHEN** owner resolution includes the self account
- **THEN** it appears under the configured `username`, indistinguishable in
  structure from an organization owner

### Requirement: TUI modes configuration

The system SHALL support a `default_mode` and a `modes` map in the configuration.
Each mode defines its TUI chrome as ordered `header` and `footer` element lists,
plus any mode-specific settings. When `modes` is omitted, the system SHALL apply
a built-in default `management` mode that reproduces the standard layout.

#### Scenario: Default management mode when modes omitted

- **WHEN** the config has no `modes` map
- **THEN** the system behaves as a single built-in `management` mode with the
  standard header/footer, and `default_mode` resolves to it

#### Scenario: Modes define per-mode chrome

- **WHEN** a mode lists `header` and `footer` element names
- **THEN** those elements, in list order, form that mode's chrome, with the
  repository list as the fixed body

#### Scenario: Unknown mode or element is rejected

- **WHEN** `default_mode` names a mode absent from `modes`, or a mode's
  `header`/`footer` names an element outside the known vocabulary
- **THEN** validation fails identifying the offending name

#### Scenario: An element may appear at most once per mode

- **WHEN** the same element appears more than once across a single mode's
  `header` and `footer`
- **THEN** validation fails identifying the duplicated element

### Requirement: Multiplex switch command configuration

The system SHALL allow a mode to declare a `switch_command`: a command template
run against a selected repository. Validation SHALL require a `switch_command`
for a mode whose behavior depends on it (the multiplex mode).

#### Scenario: switch_command is a per-mode template

- **WHEN** a mode declares a `switch_command`
- **THEN** it is stored as a template string (Go text/template) associated with
  that mode

#### Scenario: Missing switch_command for a mode that needs it

- **WHEN** a mode that requires a switch command (multiplex) omits
  `switch_command`
- **THEN** validation fails naming the mode and the missing field

