## ADDED Requirements

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
