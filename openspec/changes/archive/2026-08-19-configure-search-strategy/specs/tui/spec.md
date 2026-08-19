## MODIFIED Requirements

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
