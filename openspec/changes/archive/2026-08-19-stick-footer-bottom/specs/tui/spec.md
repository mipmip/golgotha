## ADDED Requirements

### Requirement: Footer is pinned to the bottom of the viewport

The TUI SHALL render the footer chrome at the bottom of the viewport. When the
body is shorter than the available space, the empty space SHALL appear between
the body and the footer (the header stays at the top and the body directly below
it), not below the footer.

#### Scenario: Short list pins the footer to the bottom

- **WHEN** the list has fewer rows than fit the viewport and the terminal height
  is known
- **THEN** the footer's last line is the last line of the viewport, and the gap
  is between the body and the footer

#### Scenario: Full list is unchanged

- **WHEN** the list fills the viewport
- **THEN** no padding is added and the footer remains at the bottom as before

#### Scenario: Unknown height adds no padding

- **WHEN** the terminal height is not yet known
- **THEN** no padding is applied (the view renders as before)
