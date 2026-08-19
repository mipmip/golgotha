## MODIFIED Requirements

### Requirement: Multiplex primary action clones then switches

In multiplex mode, activating a repository SHALL ensure it is cloned and then run
the mode's `switch_command` for that repository. If cloning is required it SHALL
happen first, asynchronously, showing a cancellable centered progress popup with
a determinate bar; if it fails the switch SHALL NOT run. On a successful switch
the TUI SHALL exit cleanly (status 0).

#### Scenario: Uncloned repo shows a clone popup then switches

- **WHEN** the user activates a repository that is not yet cloned in multiplex
  mode
- **THEN** the TUI shows a centered progress popup with a determinate bar while
  cloning to the templated target, and on success runs the `switch_command`

#### Scenario: Already-cloned repo switches immediately

- **WHEN** the user activates an already-cloned repository in multiplex mode
- **THEN** the TUI runs the `switch_command` without re-cloning or showing the
  clone popup

#### Scenario: Successful switch exits the program

- **WHEN** the `switch_command` runs successfully
- **THEN** the TUI quits cleanly with exit status 0

#### Scenario: Clone failure aborts the switch

- **WHEN** the required clone fails
- **THEN** the `switch_command` is not run, the popup closes, and the failure is
  reported without exiting

#### Scenario: Clone can be cancelled

- **WHEN** the user presses Esc while the clone popup is shown
- **THEN** the clone is cancelled, the popup closes, no switch runs, and the TUI
  stays open

## ADDED Requirements

### Requirement: Multiplex mode hides multiselect

In multiplex mode the repository list SHALL NOT show the multi-select checkbox,
and the selection-toggle key SHALL have no effect, because the primary action
operates on a single repository.

#### Scenario: No checkbox in multiplex

- **WHEN** the repository list is shown in multiplex mode
- **THEN** rows are rendered without the `[ ]`/`[x]` multi-select column

#### Scenario: Selection toggle is inert in multiplex

- **WHEN** the user presses the selection-toggle key in multiplex mode
- **THEN** nothing is selected and the list is unchanged

#### Scenario: Multiselect remains in management

- **WHEN** the repository list (including the combined view) is shown in
  management mode
- **THEN** the multi-select checkbox and toggle continue to work

### Requirement: Start in the combined view at launch

The TUI SHALL support launching directly into the combined cross-provider flat
view (one level deeper than the provider list), independent of the active mode.

#### Scenario: Flat-list launch

- **WHEN** the TUI is launched with the flat-list option (`--flatlist`)
- **THEN** it opens in the combined flat repository list rather than the provider
  list, and back/Esc returns to the provider list

#### Scenario: Composes with mode selection

- **WHEN** the TUI is launched with both the flat-list option and `--mode
  multiplex`
- **THEN** it opens the flat list in multiplex mode (selecting a repo clones if
  needed, switches, and exits)
