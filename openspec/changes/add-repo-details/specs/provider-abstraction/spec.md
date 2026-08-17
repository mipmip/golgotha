## ADDED Requirements

### Requirement: Repository detail and README fetch

The provider abstraction SHALL expose fetching a repository's extended details
(stars, topics, language) and its README markdown.

#### Scenario: Details are available via the interface

- **WHEN** a caller requests a repository's details
- **THEN** the provider returns its star count, topics and primary language (or
  an error)

#### Scenario: README is available via the interface

- **WHEN** a caller requests a repository's README
- **THEN** the provider returns the raw README markdown, or a not-found result
  when the repository has none
