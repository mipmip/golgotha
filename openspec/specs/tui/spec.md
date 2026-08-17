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

### Requirement: Scrolling viewport for long lists

The TUI SHALL render only the rows that fit in the available terminal height and
SHALL scroll the visible window to keep the highlighted row visible, for every
list level (providers, owners, repositories).

#### Scenario: Cursor stays visible when moving past the bottom

- **WHEN** the list is longer than the visible area and the user moves the
  cursor below the currently visible rows
- **THEN** the window scrolls so the highlighted row remains visible

#### Scenario: Cursor stays visible when moving past the top

- **WHEN** the user moves the cursor above the currently visible rows
- **THEN** the window scrolls up so the highlighted row remains visible

#### Scenario: Unknown terminal height renders all rows

- **WHEN** the terminal height is unknown (not yet reported, `height <= 0`)
- **THEN** all rows are rendered without windowing

#### Scenario: Window resets when the scope changes

- **WHEN** the user drills into or back out of a level, edits the filter, or
  refreshes
- **THEN** the scroll window resets to the top with the cursor on the first row

### Requirement: Scroll-off margin

The TUI SHALL keep a fixed margin of 2 rows between the highlighted row and the
top/bottom edge of the visible window, collapsing the margin at the start and
end of the list.

#### Scenario: Window scrolls before the cursor reaches the edge

- **WHEN** the highlighted row moves within 2 rows of the bottom of the visible
  window and more rows exist below
- **THEN** the window scrolls to preserve the 2-row margin

#### Scenario: Margin collapses at list ends

- **WHEN** the highlighted row is at or near the first or last row of the list
- **THEN** the window shows the list end without a phantom blank margin

#### Scenario: Margin is capped for small terminals

- **WHEN** the visible area is too short to hold the full margin on both sides
- **THEN** the margin is reduced so at least the highlighted row is shown

### Requirement: Paging and jump keys

The TUI SHALL provide keys to move by a screen, half a screen, and to the first
or last row, in addition to single-row movement.

#### Scenario: Page down and page up

- **WHEN** the user presses `PgDn` or `PgUp`
- **THEN** the cursor moves down or up by roughly one visible screen and the
  window follows

#### Scenario: Half-page down and up

- **WHEN** the user presses `Ctrl-D` or `Ctrl-U`
- **THEN** the cursor moves down or up by roughly half a visible screen

#### Scenario: Jump to first and last

- **WHEN** the user presses `Home` or `End`
- **THEN** the cursor jumps to the first or last row and the window follows

### Requirement: Position indicator

The TUI SHALL display the position of the visible window within the current
list.

#### Scenario: Indicator shows visible range and total

- **WHEN** a list is displayed and windowed
- **THEN** an indicator shows the visible row range and the total count
  (e.g. `41-60 of 213`)

### Requirement: Owner level lists discovered owners

The TUI SHALL populate the owner level from the cached owner index, including
owners whose repositories have not yet been fetched.

#### Scenario: Discovered owners appear before their repos are fetched

- **WHEN** `all_owners` is enabled and owners have been discovered but not all
  fetched
- **THEN** the owner level lists every discovered owner, selectable regardless of
  whether its repositories are cached yet

### Requirement: Lazy per-owner repository fetch

The TUI SHALL fetch an owner's repositories the first time the user enters it and
cache the result; subsequent visits use the cache.

#### Scenario: Fetch on first entry

- **WHEN** the user enters an owner whose repositories have not been fetched
- **THEN** the TUI fetches them, shows a loading indicator while doing so, then
  displays and caches them

#### Scenario: Cached owner is instant

- **WHEN** the user enters an owner whose repositories are already cached
- **THEN** the repositories display immediately without re-fetching

#### Scenario: Refresh re-fetches the current owner

- **WHEN** the user triggers refresh while viewing an owner
- **THEN** that owner's repositories are re-fetched and the cache updated

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

