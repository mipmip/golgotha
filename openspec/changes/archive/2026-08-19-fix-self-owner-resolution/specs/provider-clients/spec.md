## ADDED Requirements

### Requirement: Self account fetched via the authenticated endpoint

Each provider client SHALL fetch the self account's repositories via the
authenticated-user endpoint (which returns private repositories) and all other
owners via the organization/group endpoint. The client SHALL identify the self
account by comparing the requested owner to the provider's configured
`username`, not by an empty sentinel owner.

#### Scenario: Requested owner equals the configured username

- **WHEN** a fetch is requested for an owner equal to the provider's configured
  `username`
- **THEN** the client uses the authenticated-user repositories endpoint (e.g.
  GitHub `/user/repos` with owner affiliation), including private repositories

#### Scenario: Requested owner is an organization

- **WHEN** a fetch is requested for an owner different from the configured
  `username`
- **THEN** the client uses the organization/group repositories endpoint for that
  owner

#### Scenario: Fetched self repos carry the real owner login

- **WHEN** the self account's repositories are fetched
- **THEN** each returned repository is keyed by its real owner login (the
  configured `username`), so it matches how it is stored, scoped and displayed
  downstream
