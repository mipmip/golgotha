## ADDED Requirements

### Requirement: Search strategy configuration

The system SHALL support a top-level `search_strategy` field that selects how a
bare (unanchored) TUI filter term is matched: `fuzzy` (subsequence) or
`substring` (case-insensitive substring). It defaults to `fuzzy` when omitted,
preserving prior behavior.

#### Scenario: Default search strategy

- **WHEN** the config omits `search_strategy`
- **THEN** the loaded configuration uses `search_strategy=fuzzy`

#### Scenario: Substring strategy is accepted

- **WHEN** the config sets `search_strategy: substring`
- **THEN** the loaded configuration uses `search_strategy=substring`

#### Scenario: Unknown search strategy is rejected

- **WHEN** `search_strategy` is set to a value other than `fuzzy` or `substring`
- **THEN** validation fails with a clear message naming the offending value and
  the allowed values
