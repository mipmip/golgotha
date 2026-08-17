## ADDED Requirements

### Requirement: Page-aware, progress-emitting fetch

Each provider client SHALL fetch an owner's repositories page by page, emitting a
progress event per page, and determine the total page count when the API exposes
it.

#### Scenario: Total is derived from the first page

- **WHEN** the first page of an owner's repositories is fetched
- **THEN** the client derives the total page count from the provider's paging
  metadata (GitHub `Link rel="last"`, GitLab `X-Total-Pages`, Gitea
  `X-Total-Count`) when available

#### Scenario: Remaining pages fetched bounded-parallel

- **WHEN** the total page count is known and greater than one
- **THEN** the client fetches pages 2..N with at most the fixed worker cap in
  flight and merges the results, deduped by owner/name

#### Scenario: Fallback when total is unknown

- **WHEN** the provider does not expose a page total
- **THEN** the client falls back to sequential pagination, still emitting a page
  event per page

### Requirement: Non-progress fetch remains available

The system SHALL keep a simple fetch entry point that returns all repositories
without requiring an event consumer.

#### Scenario: Simple ListRepos still works

- **WHEN** a caller invokes the plain list operation
- **THEN** it returns the full repository slice as before (implemented over the
  event-emitting fetch)
