## ADDED Requirements

### Requirement: Launch-selected TUI mode

The TUI SHALL run in a mode selected at launch: the `--mode <name>` flag when
given, otherwise the configured `default_mode`, otherwise the built-in
`management` mode. The mode determines the chrome and the primary action for the
session.

#### Scenario: Flag overrides default

- **WHEN** the TUI is launched with `--mode multiplex`
- **THEN** it runs in multiplex mode regardless of `default_mode`

#### Scenario: Falls back to configured default

- **WHEN** no `--mode` flag is given
- **THEN** the TUI runs in `default_mode`, or the built-in `management` mode when
  none is configured

#### Scenario: Unknown mode name errors

- **WHEN** `--mode` names a mode that is not configured
- **THEN** the TUI exits with an actionable error naming the unknown mode

### Requirement: Mode-aware header and footer element slots

The TUI SHALL render the header and footer from the active mode's ordered
element lists, placing each named element in its region; the repository list is
always the body. Inactive conditional elements (e.g. the filter when not
filtering) SHALL be skipped, and the available list height SHALL derive from the
rendered chrome (wrap-aware). An empty list renders no chrome for that region.

#### Scenario: Elements render per active mode in order

- **WHEN** the active mode lists header/footer elements
- **THEN** they render in that region in order, with the repository list between
  them

#### Scenario: Empty slot renders no chrome

- **WHEN** the active mode's header or footer is empty (e.g. multiplex)
- **THEN** no lines are rendered for that region and the body uses the space

#### Scenario: Height tracks rendered chrome

- **WHEN** elements are added, removed, inactive, or wrap to multiple lines
- **THEN** the available list-row count reflects the actual rendered header and
  footer height

### Requirement: Multiplex primary action clones then switches

In multiplex mode, activating a repository SHALL ensure it is cloned and then run
the mode's `switch_command` for that repository. If cloning is required it SHALL
happen first with progress; if it fails the switch SHALL NOT run.

#### Scenario: Uncloned repo is cloned before switching

- **WHEN** the user activates a repository that is not yet cloned in multiplex
  mode
- **THEN** the TUI clones it to its templated target (showing progress) and, on
  success, runs the `switch_command`

#### Scenario: Already-cloned repo switches immediately

- **WHEN** the user activates an already-cloned repository in multiplex mode
- **THEN** the TUI runs the `switch_command` without re-cloning

#### Scenario: Clone failure aborts the switch

- **WHEN** the required clone fails
- **THEN** the `switch_command` is not run and the failure is reported

### Requirement: switch_command is rendered and executed shell-safely

The TUI SHALL render the active mode's `switch_command` as a Go text/template
against the selected repository — including a `Target` field holding the resolved
local clone path alongside the clone-path fields — then execute it without a
shell by splitting the rendered string into arguments using shell-quoting rules,
so repository or owner names cannot inject shell behavior.

#### Scenario: Template has access to the local path

- **WHEN** a `switch_command` references `{{.Target}}`
- **THEN** it receives the repository's resolved local clone path

#### Scenario: Names with metacharacters do not inject

- **WHEN** a repository or owner name contains spaces or shell metacharacters and
  appears in the rendered command
- **THEN** it is passed as a single literal argument, not interpreted by a shell

#### Scenario: Command result is surfaced

- **WHEN** the executed command exits non-zero
- **THEN** the TUI reports the failure rather than silently continuing
