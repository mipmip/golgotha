## ADDED Requirements

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
