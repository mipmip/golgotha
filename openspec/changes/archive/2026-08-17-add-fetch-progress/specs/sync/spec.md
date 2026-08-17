## ADDED Requirements

### Requirement: CLI fetch progress output

`skull2 sync` and `skull2 refresh` SHALL print progress derived from the fetch
event stream.

#### Scenario: Per-owner progress lines

- **WHEN** the CLI fetches repositories for owners
- **THEN** it prints line-oriented progress per owner (e.g. start, completion
  with a repo count, and any warnings), remaining cron-friendly

### Requirement: Bounded-parallel owner fetch on the CLI

The CLI SHALL fetch multiple owners concurrently within the fixed worker cap.

#### Scenario: Owners fetched concurrently

- **WHEN** many owners must be fetched
- **THEN** at most the fixed worker cap are fetched at once, and the printed
  progress still attributes each line to its owner
