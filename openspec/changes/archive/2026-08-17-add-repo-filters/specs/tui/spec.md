## ADDED Requirements

### Requirement: Interactive facet filters

The TUI SHALL provide tri-state facet filters for fork and archived status and a
value-cycle filter for visibility, applied to the currently cached repositories
and composed with the fuzzy filter.

#### Scenario: Fork facet narrows the list

- **WHEN** the user cycles the fork facet to `only` or `hide`
- **THEN** the list shows only forks, or excludes forks, respectively

#### Scenario: Archived facet narrows the list

- **WHEN** the user cycles the archived facet to `only` or `hide`
- **THEN** the list shows only archived, or excludes archived, respectively

#### Scenario: Visibility facet cycles values

- **WHEN** the user cycles the visibility facet through `all` → `public` →
  `private` → `internal`
- **THEN** the list is limited to repositories of the selected visibility (all
  when `all`)

#### Scenario: Facets compose with fuzzy search

- **WHEN** one or more facets are active and a fuzzy query is entered
- **THEN** the visible repositories satisfy every active facet AND the fuzzy
  query

#### Scenario: Active facets are visible

- **WHEN** any facet is not `all`
- **THEN** the UI indicates the active facet states (e.g. `fork:hide vis:private`)

### Requirement: Hint when a facet cannot match cached data

The TUI SHALL indicate when a facet selection matches nothing because that data
was excluded before caching.

#### Scenario: Only-archived with archived not cached

- **WHEN** the archived facet is set to `only` but archived repositories were not
  cached (config excluded them)
- **THEN** the UI shows a hint explaining that archived repos are not cached,
  rather than an unexplained empty list

### Requirement: Filtering interacts with the viewport

The TUI SHALL recompute the scrolling window against the filtered list.

#### Scenario: Changing a facet resets the window

- **WHEN** a facet selection changes
- **THEN** the list re-filters and the scroll window resets to the top with the
  cursor on the first row
