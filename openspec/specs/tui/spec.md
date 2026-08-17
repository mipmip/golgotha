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

### Requirement: Interactive facet filters

The TUI SHALL provide tri-state facet filters for fork and archived status and a
value-cycle filter for visibility, applied to the currently cached repositories
and composed with the fuzzy filter.

#### Scenario: Fork facet narrows the list

- **WHEN** the user cycles the fork facet to `only` or `hide`
- **THEN** the list shows only forks, or excludes forks, respectively

#### Scenario: Archived facet narrows the list

- **WHEN** the user cycles the archived facet to `only` or `hide`
- **THEN** the list shows only archived, or excludes archived, respectively

#### Scenario: Visibility facet cycles values

- **WHEN** the user cycles the visibility facet through `all` → `public` →
  `private` → `internal`
- **THEN** the list is limited to repositories of the selected visibility (all
  when `all`)

#### Scenario: Facets compose with fuzzy search

- **WHEN** one or more facets are active and a fuzzy query is entered
- **THEN** the visible repositories satisfy every active facet AND the fuzzy
  query

#### Scenario: Active facets are visible

- **WHEN** any facet is not `all`
- **THEN** the UI indicates the active facet states (e.g. `fork:hide vis:private`)

### Requirement: Hint when a facet cannot match cached data

The TUI SHALL indicate when a facet selection matches nothing because that data
was excluded before caching.

#### Scenario: Only-archived with archived not cached

- **WHEN** the archived facet is set to `only` but archived repositories were not
  cached (config excluded them)
- **THEN** the UI shows a hint explaining that archived repos are not cached,
  rather than an unexplained empty list

### Requirement: Filtering interacts with the viewport

The TUI SHALL recompute the scrolling window against the filtered list.

#### Scenario: Changing a facet resets the window

- **WHEN** a facet selection changes
- **THEN** the list re-filters and the scroll window resets to the top with the
  cursor on the first row

### Requirement: Enter opens the repository detail view

At the repository level the TUI SHALL open a detail view on Enter; cloning is
triggered only by `c`.

#### Scenario: Enter drills into details

- **WHEN** the user presses Enter on a repository row
- **THEN** the detail view for that repository opens (Enter no longer clones)

#### Scenario: Clone is on c

- **WHEN** the user presses `c` on a repository row or selection
- **THEN** the repository (or selection) is cloned as before

### Requirement: Detail view content and navigation

The detail view SHALL show the repository's metadata and its rendered README, and
allow scrolling, opening, cloning, refreshing and going back.

#### Scenario: Metadata and README shown

- **WHEN** the detail view is displayed
- **THEN** it shows description, stars, topics, language, last-updated and
  visibility, plus the scrollable rendered README

#### Scenario: Loading indicator while fetching

- **WHEN** details are being fetched on first open
- **THEN** a loading indicator is shown until they arrive

#### Scenario: Back returns to the repo list

- **WHEN** the user presses Esc in the detail view
- **THEN** the view returns to the repository list at the prior position

