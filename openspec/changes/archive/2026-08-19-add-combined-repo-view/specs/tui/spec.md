## ADDED Requirements

### Requirement: Combined all-repositories view

The TUI SHALL provide a combined, cross-provider view that lists every cached
repository across all providers and owners in one flat list, reachable from an
"All repositories" entry in the provider list. The view SHALL reuse the existing repository
row behavior (fuzzy filter, facet filters, sort, single and bulk clone, and
detail) applied across the whole set. Leaving the view SHALL return to the
provider list.

#### Scenario: Enter the combined view

- **WHEN** the user selects the "All repositories" entry in the provider list
- **THEN** the TUI shows a flat list of every cached repository across all
  providers and owners, and back/Esc returns to the provider list

#### Scenario: Rows are disambiguated by provider and owner

- **WHEN** the combined list is displayed
- **THEN** each row is labeled so its provider and owner are unambiguous
  (a provider short code plus `owner/name`)

#### Scenario: Filters and sort apply across the whole set

- **WHEN** the user applies the fuzzy filter, a facet filter, or a sort while in
  the combined view
- **THEN** it applies across all providers and owners, not a single provider or
  owner

#### Scenario: Actions work regardless of provider

- **WHEN** the user opens the detail view for, or clones (single or bulk), a row
  in the combined view
- **THEN** the action uses that row's own provider and resolved target,
  independent of any drilled-down selection

### Requirement: Combined view reports cache completeness

The combined view SHALL indicate how complete the underlying cache is, so a
partially-fetched cache is visible rather than silently incomplete.

#### Scenario: Completeness is shown

- **WHEN** the combined view is displayed
- **THEN** it shows the number of repositories and how many owners are loaded out
  of the total known owners (e.g. "3/8 owners loaded")

#### Scenario: Incomplete cache hints at refresh

- **WHEN** some known owners have not been fetched
- **THEN** the view indicates that a refresh will load the remaining owners

### Requirement: Full cross-provider refresh from the combined view

The combined view SHALL offer a full refresh that re-fetches every owner across
every provider, updating the cache, with progress feedback, so the user can make
the overview complete and current.

#### Scenario: Refresh fetches all providers and owners

- **WHEN** the user triggers refresh in the combined view
- **THEN** the TUI re-fetches every owner across every provider, shows progress
  while doing so, and updates the combined list and the cache on completion

#### Scenario: Per-provider refresh is atomic

- **WHEN** a provider's refresh during a combined refresh fails or is interrupted
- **THEN** that provider's previously cached repositories remain intact (each
  provider commits only on a complete refresh)
