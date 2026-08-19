## ADDED Requirements

### Requirement: Clone can emit progress

The clone engine SHALL offer a clone operation that emits progress events
(a percentage and a phase label) as the clone proceeds, so an interactive caller
can render a determinate progress bar. The existing non-progress clone used by
`hup sync` is unaffected.

#### Scenario: Progress is derived from git

- **WHEN** a progress-emitting clone runs
- **THEN** it invokes `git clone --progress` and parses git's progress output
  (e.g. "Receiving objects: N%") into percentage/phase events

#### Scenario: Completion and failure are reported

- **WHEN** a progress-emitting clone finishes or fails
- **THEN** the caller receives a terminal result (success with the resolved
  target, or an error) after the progress events

#### Scenario: Cancellation stops the clone

- **WHEN** the caller cancels the context of a progress-emitting clone
- **THEN** the underlying git process is stopped and the operation returns
  promptly without leaving a completed clone reported
