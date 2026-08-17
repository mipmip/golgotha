## ADDED Requirements

### Requirement: Lazy repository detail fetch

The system SHALL fetch a repository's extended details (stars, topics, language)
and its README on first open, and reuse them thereafter.

#### Scenario: Details fetched on first open

- **WHEN** a repository's detail view is opened and no detail cache exists
- **THEN** the system fetches stars, topics, language and the README, then
  displays them

#### Scenario: Cached details reused

- **WHEN** a repository's detail view is opened and a detail cache exists
- **THEN** it is shown without re-fetching

#### Scenario: Manual refresh

- **WHEN** the user triggers refresh in the detail view
- **THEN** the repository's details and README are re-fetched and the detail
  cache updated

### Requirement: Separate detail cache

The system SHALL store repository details in a per-repository cache file separate
from the list cache, holding raw README markdown.

#### Scenario: Detail cache is separate from the list cache

- **WHEN** details are cached
- **THEN** they are written under `details/<provider>/<owner>__<repo>.json` with
  a fetch timestamp and raw README markdown, and the main `<provider>.json` list
  cache is not modified

### Requirement: Rendered README

The system SHALL render the README markdown for display, at the current width.

#### Scenario: Markdown is rendered

- **WHEN** a README is displayed
- **THEN** it is rendered from the stored raw markdown to styled terminal text at
  the current view width

### Requirement: Graceful offline detail

The system SHALL degrade gracefully when detail fetching fails.

#### Scenario: Fetch fails with no cache

- **WHEN** the detail fetch fails and no detail cache exists
- **THEN** the view shows the already-known metadata and a "README unavailable"
  note, without an error screen
