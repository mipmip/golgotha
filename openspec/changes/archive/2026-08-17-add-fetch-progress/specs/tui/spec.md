## ADDED Requirements

### Requirement: Fetch progress on owner entry

The TUI SHALL show textual progress with a bar while lazily fetching an owner's
repositories on entry.

#### Scenario: Spinner then determinate bar

- **WHEN** the user enters an unfetched owner
- **THEN** the TUI shows an indeterminate spinner until the total is known, then
  a determinate progress bar and a line describing the current step (e.g.
  "fetching <owner> page i/n — N repos")

#### Scenario: Progress advances per page

- **WHEN** each page of the owner completes
- **THEN** the bar and the textual description update to reflect pages done and
  repos fetched so far

### Requirement: Cancel the current fetch

The TUI SHALL let the user cancel the in-flight fetch.

#### Scenario: Esc cancels and backs out

- **WHEN** the user presses Esc during an owner fetch
- **THEN** the fetch is cancelled, the owner is left unfetched, and the UI
  returns to the owner list without caching partial results

#### Scenario: Partial results are transient

- **WHEN** a fetch is cancelled or fails partway
- **THEN** any partial repositories are not presented as the owner's complete set
  and are not cached
