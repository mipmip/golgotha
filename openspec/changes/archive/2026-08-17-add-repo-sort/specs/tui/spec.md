## ADDED Requirements

### Requirement: Repository list sorting

The system SHALL let the user sort the repository list by name or by last-updated
time, in ascending or descending direction, while preserving the provider/cache
fetch order as the default and as a selectable state.

#### Scenario: Default is fetch order

- **WHEN** the repository list is first shown and the user has not chosen a sort
- **THEN** repositories appear in the provider/cache fetch order (no sort applied)

#### Scenario: Cycle the sort key

- **WHEN** the user presses `s`
- **THEN** the active sort key advances through the cycle `none → name →
  last-updated → none`, and the visible list reorders accordingly (with `none`
  restoring fetch order)

#### Scenario: Sort by name

- **WHEN** the active sort key is name in ascending direction
- **THEN** repositories are ordered case-insensitively by name A→Z

#### Scenario: Sort by last-updated

- **WHEN** the active sort key is last-updated in descending direction
- **THEN** repositories are ordered by `UpdatedAt` with the most recently updated
  first

#### Scenario: Reverse the direction

- **WHEN** a sort key other than `none` is active and the user presses `S`
- **THEN** the sort direction toggles between ascending and descending and the
  visible list reorders accordingly

#### Scenario: Sort applies after filtering

- **WHEN** a fuzzy filter query is active and a sort key is selected
- **THEN** the sort orders only the filtered (visible) repositories, not the
  hidden ones

#### Scenario: Active sort is shown in the footer

- **WHEN** a sort key other than `none` is active
- **THEN** the footer help bar indicates the active sort key and direction
