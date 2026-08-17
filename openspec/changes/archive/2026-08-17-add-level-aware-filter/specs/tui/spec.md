## ADDED Requirements

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

While filtering, Enter SHALL act on the highlighted item using the current
level's normal behavior.

#### Scenario: Enter on a filtered owner drills in

- **WHEN** the owner list is filtered and the user presses Enter on a match
- **THEN** the app drills into that owner's repositories

### Requirement: Filter clears on level change

The filter SHALL be cleared when the navigation level changes.

#### Scenario: Drilling in starts unfiltered

- **WHEN** the user drills into a level while a filter is active
- **THEN** the new level is shown unfiltered

#### Scenario: Going back starts unfiltered

- **WHEN** the user navigates back while a filter is active
- **THEN** the filter is cleared
