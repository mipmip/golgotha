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

The system SHALL provide a global filter across repositories that parses the
query as an fzf-style extended-search expression: whitespace-separated AND
terms, each term optionally negated (`!`) and optionally anchored as a prefix
(`^`), suffix (`$`), or exact-substring/fuzzy toggle (`'`). Within an AND group,
terms separated by `|` form an OR sub-group (`a b | c` means `a AND (b OR c)`).
A bare term (no operator) is matched according to the configured
`search_strategy`: subsequence when `fuzzy`, case-insensitive substring when
`substring`. The `'` prefix toggles a term to the opposite of the configured
strategy. All matching is case-insensitive. An empty query matches everything.

The extended matcher SHALL apply uniformly at every navigation level (providers,
owners, repos) and in the flat all-repositories view, replacing the previous
subsequence-only behavior.

#### Scenario: Filter narrows the list

- **WHEN** the user activates the filter and types a query
- **THEN** the visible repositories are limited to matches of the parsed query

#### Scenario: Bare term follows the configured strategy

- **WHEN** `search_strategy` is `substring` and the user types `nivis`
- **THEN** only repositories whose matched string contains the literal substring
  `nivis` (case-insensitive) are shown, excluding mere subsequence matches

#### Scenario: Quote toggles to the opposite strategy

- **WHEN** `search_strategy` is `fuzzy` and the user types `'nivis`
- **THEN** only literal-substring matches of `nivis` are shown; and **WHEN**
  `search_strategy` is `substring` and the user types `'nvs`, subsequence
  matches of `nvs` are shown

#### Scenario: Prefix and suffix anchors

- **WHEN** the user types `^foo`
- **THEN** only matched strings that start with `foo` are shown; and **WHEN** the
  user types `bar$`, only matched strings that end with `bar` are shown

#### Scenario: Negation excludes matches

- **WHEN** the user types `!archived`
- **THEN** repositories whose matched string matches `archived` are excluded from
  the visible list

#### Scenario: Multiple terms AND together

- **WHEN** the user types `foo bar`
- **THEN** only repositories matching both `foo` and `bar` are shown

#### Scenario: OR within an AND group

- **WHEN** the user types `foo bar | baz`
- **THEN** only repositories matching `foo` AND (`bar` OR `baz`) are shown

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

### Requirement: Repository list sorting

The system SHALL let the user sort the repository list by name or by last-updated
time, in ascending or descending direction, while preserving the provider/cache
fetch order as the default and as a selectable state.

#### Scenario: Default is fetch order

- **WHEN** the repository list is first shown and the user has not chosen a sort
- **THEN** repositories appear in the provider/cache fetch order (no sort applied)

#### Scenario: Cycle the sort key

- **WHEN** the user presses `s`
- **THEN** the active sort key advances through the cycle `none → name →
  last-updated → none`, and the visible list reorders accordingly (with `none`
  restoring fetch order)

#### Scenario: Sort by name

- **WHEN** the active sort key is name in ascending direction
- **THEN** repositories are ordered case-insensitively by name A→Z

#### Scenario: Sort by last-updated

- **WHEN** the active sort key is last-updated in descending direction
- **THEN** repositories are ordered by `UpdatedAt` with the most recently updated
  first

#### Scenario: Reverse the direction

- **WHEN** a sort key other than `none` is active and the user presses `S`
- **THEN** the sort direction toggles between ascending and descending and the
  visible list reorders accordingly

#### Scenario: Sort applies after filtering

- **WHEN** a fuzzy filter query is active and a sort key is selected
- **THEN** the sort orders only the filtered (visible) repositories, not the
  hidden ones

#### Scenario: Active sort is shown in the footer

- **WHEN** a sort key other than `none` is active
- **THEN** the footer help bar indicates the active sort key and direction

### Requirement: Filter scopes to the current navigation level

The fuzzy filter SHALL narrow the items of the current navigation level, not
always the repository list.

#### Scenario: Filtering at the owners level filters organizations

- **WHEN** the user is at the owners level and types a filter query
- **THEN** the owner/organization list is narrowed to fuzzy matches (the view
  does not switch to repositories)

#### Scenario: Filtering at the providers level filters providers

- **WHEN** the user is at the providers level and types a filter query
- **THEN** the provider list is narrowed to fuzzy matches

#### Scenario: Filtering at the repos level filters repositories

- **WHEN** the user is at the repos level and types a filter query
- **THEN** the repository list is narrowed to fuzzy matches (as before)

#### Scenario: Match is against the raw name

- **WHEN** an owner label carries a decoration such as "(not fetched)"
- **THEN** the fuzzy query matches the raw owner name, not the decoration

### Requirement: Enter drills the filtered item at its level

While filtering, a single Enter SHALL act on the highlighted item using the
current level's normal behavior (drilling into a provider or owner, or opening /
switching a repository), then leave filter-input mode.

#### Scenario: Enter on a filtered owner drills in

- **WHEN** the owner list is filtered and the user presses Enter on a match
- **THEN** the app drills into that owner's repositories in a single press

### Requirement: Filter clears on level change

The filter SHALL be cleared when the navigation level changes.

#### Scenario: Drilling in starts unfiltered

- **WHEN** the user drills into a level while a filter is active
- **THEN** the new level is shown unfiltered

#### Scenario: Going back starts unfiltered

- **WHEN** the user navigates back while a filter is active
- **THEN** the filter is cleared

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

### Requirement: Self account is an ordinary owner, pinned and distinguished

The TUI SHALL present the self account as an ordinary owner identified by its
real login (the configured `username`). It SHALL pin the self account first in
the owner list and SHALL visually distinguish it from organization owners.
Entering the self owner SHALL scope the repository list to only that account's
repositories.

#### Scenario: Self owner is pinned first and distinguished

- **WHEN** the owner level is displayed for a provider
- **THEN** the self account (the configured `username`) appears first in the
  owner list and is visually distinguished (e.g. a distinct color) from
  organization owners

#### Scenario: Self owner label is the real login

- **WHEN** the self account is shown in the owner list
- **THEN** its label is the configured `username`, not a generic placeholder
  such as "(your account)"

#### Scenario: Entering the self owner scopes to its own repos

- **WHEN** the user enters the self owner
- **THEN** only repositories whose owner equals the configured `username` are
  shown, not every repository for the provider

### Requirement: Launch-selected TUI mode

The TUI SHALL run in a mode selected at launch: the `--mode <name>` flag when
given, otherwise the configured `default_mode`, otherwise the built-in
`management` mode. The mode determines the chrome and the primary action for the
session.

#### Scenario: Flag overrides default

- **WHEN** the TUI is launched with `--mode multiplex`
- **THEN** it runs in multiplex mode regardless of `default_mode`

#### Scenario: Falls back to configured default

- **WHEN** no `--mode` flag is given
- **THEN** the TUI runs in `default_mode`, or the built-in `management` mode when
  none is configured

#### Scenario: Unknown mode name errors

- **WHEN** `--mode` names a mode that is not configured
- **THEN** the TUI exits with an actionable error naming the unknown mode

### Requirement: Mode-aware header and footer element slots

The TUI SHALL render the header and footer from the active mode's ordered
element lists, placing each named element in its region; the repository list is
always the body. Inactive conditional elements (e.g. the filter when not
filtering) SHALL be skipped, and the available list height SHALL derive from the
rendered chrome (wrap-aware). An empty list renders no chrome for that region.

#### Scenario: Elements render per active mode in order

- **WHEN** the active mode lists header/footer elements
- **THEN** they render in that region in order, with the repository list between
  them

#### Scenario: Empty slot renders no chrome

- **WHEN** the active mode's header or footer is empty (e.g. multiplex)
- **THEN** no lines are rendered for that region and the body uses the space

#### Scenario: Height tracks rendered chrome

- **WHEN** elements are added, removed, inactive, or wrap to multiple lines
- **THEN** the available list-row count reflects the actual rendered header and
  footer height

### Requirement: Multiplex primary action clones then switches

In multiplex mode, activating a repository SHALL ensure it is cloned and then run
the mode's `switch_command` for that repository. If cloning is required it SHALL
happen first, asynchronously, showing a cancellable centered progress popup with
a determinate bar; if it fails the switch SHALL NOT run. On a successful switch
the TUI SHALL exit cleanly (status 0).

#### Scenario: Uncloned repo shows a clone popup then switches

- **WHEN** the user activates a repository that is not yet cloned in multiplex
  mode
- **THEN** the TUI shows a centered progress popup with a determinate bar while
  cloning to the templated target, and on success runs the `switch_command`

#### Scenario: Already-cloned repo switches immediately

- **WHEN** the user activates an already-cloned repository in multiplex mode
- **THEN** the TUI runs the `switch_command` without re-cloning or showing the
  clone popup

#### Scenario: Successful switch exits the program

- **WHEN** the `switch_command` runs successfully
- **THEN** the TUI quits cleanly with exit status 0

#### Scenario: Clone failure aborts the switch

- **WHEN** the required clone fails
- **THEN** the `switch_command` is not run, the popup closes, and the failure is
  reported without exiting

#### Scenario: Clone can be cancelled

- **WHEN** the user presses Esc while the clone popup is shown
- **THEN** the clone is cancelled, the popup closes, no switch runs, and the TUI
  stays open

### Requirement: switch_command is rendered and executed shell-safely

The TUI SHALL render the active mode's `switch_command` as a Go text/template
against the selected repository — including a `Target` field holding the resolved
local clone path alongside the clone-path fields — then execute it without a
shell by splitting the rendered string into arguments using shell-quoting rules,
so repository or owner names cannot inject shell behavior.

#### Scenario: Template has access to the local path

- **WHEN** a `switch_command` references `{{.Target}}`
- **THEN** it receives the repository's resolved local clone path

#### Scenario: Names with metacharacters do not inject

- **WHEN** a repository or owner name contains spaces or shell metacharacters and
  appears in the rendered command
- **THEN** it is passed as a single literal argument, not interpreted by a shell

#### Scenario: Command result is surfaced

- **WHEN** the executed command exits non-zero
- **THEN** the TUI reports the failure rather than silently continuing

### Requirement: Combined all-repositories view

The TUI SHALL provide a combined, cross-provider view that lists every cached
repository across all providers and owners in one flat list, reachable from an
"All repositories" entry in the provider list. The view SHALL reuse the existing repository
row behavior (fuzzy filter, facet filters, sort, single and bulk clone, and
detail) applied across the whole set. Leaving the view SHALL return to the
provider list.

#### Scenario: Enter the combined view

- **WHEN** the user selects the "All repositories" entry in the provider list
- **THEN** the TUI shows a flat list of every cached repository across all
  providers and owners, and back/Esc returns to the provider list

#### Scenario: Rows are disambiguated by provider and owner

- **WHEN** the combined list is displayed
- **THEN** each row is labeled so its provider and owner are unambiguous
  (a provider short code plus `owner/name`)

#### Scenario: Filters and sort apply across the whole set

- **WHEN** the user applies the fuzzy filter, a facet filter, or a sort while in
  the combined view
- **THEN** it applies across all providers and owners, not a single provider or
  owner

#### Scenario: Actions work regardless of provider

- **WHEN** the user opens the detail view for, or clones (single or bulk), a row
  in the combined view
- **THEN** the action uses that row's own provider and resolved target,
  independent of any drilled-down selection

### Requirement: Combined view reports cache completeness

The combined view SHALL indicate how complete the underlying cache is, so a
partially-fetched cache is visible rather than silently incomplete.

#### Scenario: Completeness is shown

- **WHEN** the combined view is displayed
- **THEN** it shows the number of repositories and how many owners are loaded out
  of the total known owners (e.g. "3/8 owners loaded")

#### Scenario: Incomplete cache hints at refresh

- **WHEN** some known owners have not been fetched
- **THEN** the view indicates that a refresh will load the remaining owners

### Requirement: Full cross-provider refresh from the combined view

The combined view SHALL offer a full refresh that re-fetches every owner across
every provider, updating the cache, with progress feedback, so the user can make
the overview complete and current.

#### Scenario: Refresh fetches all providers and owners

- **WHEN** the user triggers refresh in the combined view
- **THEN** the TUI re-fetches every owner across every provider, shows progress
  while doing so, and updates the combined list and the cache on completion

#### Scenario: Per-provider refresh is atomic

- **WHEN** a provider's refresh during a combined refresh fails or is interrupted
- **THEN** that provider's previously cached repositories remain intact (each
  provider commits only on a complete refresh)

### Requirement: Navigate the list while filtering

While the fuzzy filter is being typed, the TUI SHALL let the user move the
selection without leaving filter-input mode. Vertical navigation keys move the
selection; printable keys continue to edit the query. The selection SHALL reset
to the top only when the query text changes, so navigating does not lose the
highlight.

#### Scenario: Arrows move the selection while filtering

- **WHEN** the filter input is active and the user presses Up or Down (or PgUp /
  PgDn, or Ctrl+P / Ctrl+N)
- **THEN** the highlighted row moves within the filtered list while the filter
  input stays focused and the query is unchanged

#### Scenario: Typing narrows and resets to the top

- **WHEN** the user changes the filter query text
- **THEN** the list re-filters and the selection resets to the first row

#### Scenario: Navigation preserves the highlight

- **WHEN** the user moves the selection and then presses a navigation key again
  without changing the query
- **THEN** the highlight is preserved and moves from its current position (it is
  not reset to the top)

#### Scenario: Letter keys still type

- **WHEN** the user presses a printable key such as `j` or `k` while filtering
- **THEN** it is appended to the query (it does not move the selection)

### Requirement: Multiplex mode hides multiselect

In multiplex mode the repository list SHALL NOT show the multi-select checkbox,
and the selection-toggle key SHALL have no effect, because the primary action
operates on a single repository.

#### Scenario: No checkbox in multiplex

- **WHEN** the repository list is shown in multiplex mode
- **THEN** rows are rendered without the `[ ]`/`[x]` multi-select column

#### Scenario: Selection toggle is inert in multiplex

- **WHEN** the user presses the selection-toggle key in multiplex mode
- **THEN** nothing is selected and the list is unchanged

#### Scenario: Multiselect remains in management

- **WHEN** the repository list (including the combined view) is shown in
  management mode
- **THEN** the multi-select checkbox and toggle continue to work

### Requirement: Start in the combined view at launch

The TUI SHALL support launching directly into the combined cross-provider flat
view (one level deeper than the provider list), independent of the active mode.

#### Scenario: Flat-list launch

- **WHEN** the TUI is launched with the flat-list option (`--flatlist`)
- **THEN** it opens in the combined flat repository list rather than the provider
  list, and back/Esc returns to the provider list

#### Scenario: Composes with mode selection

- **WHEN** the TUI is launched with both the flat-list option and `--mode
  multiplex`
- **THEN** it opens the flat list in multiplex mode (selecting a repo clones if
  needed, switches, and exits)

### Requirement: Footer is pinned to the bottom of the viewport

The TUI SHALL render the footer chrome at the bottom of the viewport. When the
body is shorter than the available space, the empty space SHALL appear between
the body and the footer (the header stays at the top and the body directly below
it), not below the footer.

#### Scenario: Short list pins the footer to the bottom

- **WHEN** the list has fewer rows than fit the viewport and the terminal height
  is known
- **THEN** the footer's last line is the last line of the viewport, and the gap
  is between the body and the footer

#### Scenario: Full list is unchanged

- **WHEN** the list fills the viewport
- **THEN** no padding is added and the footer remains at the bottom as before

#### Scenario: Unknown height adds no padding

- **WHEN** the terminal height is not yet known
- **THEN** no padding is applied (the view renders as before)

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

