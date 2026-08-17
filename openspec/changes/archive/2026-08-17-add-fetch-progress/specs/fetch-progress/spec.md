## ADDED Requirements

### Requirement: Progress event stream

The system SHALL provide a progress-event stream that a repository fetch emits
and that any frontend can consume, decoupling work from presentation.

#### Scenario: Fetch emits lifecycle events

- **WHEN** a repository fetch for an owner runs
- **THEN** it emits a start event, a page event per fetched page (with page
  number, total pages when known, and repos so far), and a terminal event
  (done with a count, failed with an error, or canceled)

#### Scenario: Same stream feeds multiple frontends

- **WHEN** the same fetch runs under the TUI or the CLI
- **THEN** both render from the identical event stream (bar/text vs printed
  lines), and tests can assert on the emitted sequence

### Requirement: Bounded parallelism

The system SHALL bound fetch concurrency to a fixed worker cap.

#### Scenario: Pages fan out after the first

- **WHEN** an owner's total page count is known after the first page
- **THEN** the remaining pages are fetched with at most the fixed worker cap in
  flight

#### Scenario: Owners fan out on the CLI

- **WHEN** the CLI fetches multiple owners
- **THEN** at most the fixed worker cap of owners are fetched concurrently

### Requirement: Cancellation

The system SHALL support cancelling an in-flight fetch via context.

#### Scenario: Cancel stops work and emits canceled

- **WHEN** a running fetch is cancelled
- **THEN** in-flight work stops promptly and a canceled event is emitted for the
  affected owner
