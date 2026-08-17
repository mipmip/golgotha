## ADDED Requirements

### Requirement: Completed lazy fetch displays without restart

When a lazy per-owner fetch completes successfully, the TUI SHALL display that
owner's repositories immediately, without requiring a restart.

#### Scenario: Repos appear right after fetching a new owner

- **WHEN** the user enters a not-yet-fetched owner and the fetch completes
- **THEN** the owner's repositories are shown in the current session (the cache
  is committed before the completion is handled)

#### Scenario: Fetch always reaches a terminal state

- **WHEN** a lazy fetch ends — success, failure, or cancellation
- **THEN** the TUI leaves the "fetching…" state (a terminal event is always
  delivered)
