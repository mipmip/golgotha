# tui Specification

## Purpose
TBD - created by archiving change add-tui. Update Purpose after archive.
## Requirements
### Requirement: Hierarchic browsing

The system SHALL let the user navigate provider → owner → repositories, reading
from the cache.

#### Scenario: Drill down and back

- **WHEN** the user selects a provider then an owner
- **THEN** the repositories for that owner are shown, and the user can navigate
  back up the hierarchy

### Requirement: Fuzzy filter

The system SHALL provide a global fuzzy filter across repositories.

#### Scenario: Filter narrows the list

- **WHEN** the user activates the filter and types a query
- **THEN** the visible repositories are limited to fuzzy matches

### Requirement: Clone from the TUI

The system SHALL clone the selected repository, or all multi-selected
repositories, to their templated target paths.

#### Scenario: Single clone

- **WHEN** the user selects a repository and triggers clone
- **THEN** it is cloned to its resolved target and the row shows a cloned status

#### Scenario: Bulk clone

- **WHEN** the user multi-selects repositories and triggers clone
- **THEN** each is cloned and progress is shown

### Requirement: Open in browser and refresh

The system SHALL open a repository's web URL and refresh the current provider's
cache on demand.

#### Scenario: Open in browser

- **WHEN** the user triggers open on a repository
- **THEN** its `web_url` is opened in the default browser

#### Scenario: Refresh

- **WHEN** the user triggers refresh
- **THEN** the current provider's cache is re-fetched and the list updates

