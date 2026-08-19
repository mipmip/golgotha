## ADDED Requirements

### Requirement: Prefetch details on navigation

At the repository level, the TUI SHALL prefetch the highlighted repository's
details in the background once the cursor settles on it, so opening the detail
view is instant when the prefetch has completed. Prefetching SHALL be debounced
(rapid cursor movement does not fire a fetch per row), at most one prefetch runs
at a time, a superseded prefetch is cancelled, and an already-cached repository
is not prefetched.

#### Scenario: Settling on a repo warms its details

- **WHEN** the cursor settles on a repository row whose details are not cached
- **THEN** a background prefetch fetches and caches its details, so a subsequent
  Enter opens the detail view from cache without a network wait

#### Scenario: Rapid scrolling does not fetch every row

- **WHEN** the cursor moves quickly across several rows
- **THEN** prefetches for intermediate rows are debounced/cancelled and only the
  row the cursor settles on is prefetched

#### Scenario: Prefetch result does not disturb the current view

- **WHEN** a background prefetch completes for a repository other than the one
  currently open
- **THEN** the visible view is not changed by the prefetch (it only warms the
  cache)
