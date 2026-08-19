## ADDED Requirements

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
