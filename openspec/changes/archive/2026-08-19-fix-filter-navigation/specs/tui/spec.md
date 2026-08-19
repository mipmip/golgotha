## ADDED Requirements

### Requirement: Navigate the list while filtering

While the fuzzy filter is being typed, the TUI SHALL let the user move the
selection without leaving filter-input mode. Vertical navigation keys move the
selection; printable keys continue to edit the query. The selection SHALL reset
to the top only when the query text changes, so navigating does not lose the
highlight.

#### Scenario: Arrows move the selection while filtering

- **WHEN** the filter input is active and the user presses Up or Down (or PgUp /
  PgDn, or Ctrl+P / Ctrl+N)
- **THEN** the highlighted row moves within the filtered list while the filter
  input stays focused and the query is unchanged

#### Scenario: Typing narrows and resets to the top

- **WHEN** the user changes the filter query text
- **THEN** the list re-filters and the selection resets to the first row

#### Scenario: Navigation preserves the highlight

- **WHEN** the user moves the selection and then presses a navigation key again
  without changing the query
- **THEN** the highlight is preserved and moves from its current position (it is
  not reset to the top)

#### Scenario: Letter keys still type

- **WHEN** the user presses a printable key such as `j` or `k` while filtering
- **THEN** it is appended to the query (it does not move the selection)

## MODIFIED Requirements

### Requirement: Enter drills the filtered item at its level

While filtering, a single Enter SHALL act on the highlighted item using the
current level's normal behavior (drilling into a provider or owner, or opening /
switching a repository), then leave filter-input mode.

#### Scenario: Enter on a filtered owner drills in

- **WHEN** the owner list is filtered and the user presses Enter on a match
- **THEN** the app drills into that owner's repositories in a single press
