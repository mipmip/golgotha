## ADDED Requirements

### Requirement: Scrolling viewport for long lists

The TUI SHALL render only the rows that fit in the available terminal height and
SHALL scroll the visible window to keep the highlighted row visible, for every
list level (providers, owners, repositories).

#### Scenario: Cursor stays visible when moving past the bottom

- **WHEN** the list is longer than the visible area and the user moves the
  cursor below the currently visible rows
- **THEN** the window scrolls so the highlighted row remains visible

#### Scenario: Cursor stays visible when moving past the top

- **WHEN** the user moves the cursor above the currently visible rows
- **THEN** the window scrolls up so the highlighted row remains visible

#### Scenario: Unknown terminal height renders all rows

- **WHEN** the terminal height is unknown (not yet reported, `height <= 0`)
- **THEN** all rows are rendered without windowing

#### Scenario: Window resets when the scope changes

- **WHEN** the user drills into or back out of a level, edits the filter, or
  refreshes
- **THEN** the scroll window resets to the top with the cursor on the first row

### Requirement: Scroll-off margin

The TUI SHALL keep a fixed margin of 2 rows between the highlighted row and the
top/bottom edge of the visible window, collapsing the margin at the start and
end of the list.

#### Scenario: Window scrolls before the cursor reaches the edge

- **WHEN** the highlighted row moves within 2 rows of the bottom of the visible
  window and more rows exist below
- **THEN** the window scrolls to preserve the 2-row margin

#### Scenario: Margin collapses at list ends

- **WHEN** the highlighted row is at or near the first or last row of the list
- **THEN** the window shows the list end without a phantom blank margin

#### Scenario: Margin is capped for small terminals

- **WHEN** the visible area is too short to hold the full margin on both sides
- **THEN** the margin is reduced so at least the highlighted row is shown

### Requirement: Paging and jump keys

The TUI SHALL provide keys to move by a screen, half a screen, and to the first
or last row, in addition to single-row movement.

#### Scenario: Page down and page up

- **WHEN** the user presses `PgDn` or `PgUp`
- **THEN** the cursor moves down or up by roughly one visible screen and the
  window follows

#### Scenario: Half-page down and up

- **WHEN** the user presses `Ctrl-D` or `Ctrl-U`
- **THEN** the cursor moves down or up by roughly half a visible screen

#### Scenario: Jump to first and last

- **WHEN** the user presses `Home` or `End`
- **THEN** the cursor jumps to the first or last row and the window follows

### Requirement: Position indicator

The TUI SHALL display the position of the visible window within the current
list.

#### Scenario: Indicator shows visible range and total

- **WHEN** a list is displayed and windowed
- **THEN** an indicator shows the visible row range and the total count
  (e.g. `41-60 of 213`)
