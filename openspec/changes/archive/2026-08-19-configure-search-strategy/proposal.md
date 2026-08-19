## Why

The TUI filter uses subsequence ("fuzzy") matching, where `nivis` matches any
string containing `n…i…v…i…s` in order. This floods results with irrelevant
matches when the user wants a literal substring. Users have no way to say "just
show me repos whose org or name actually contains this text."

## What Changes

- Replace the single subsequence matcher with an **fzf-style extended-search**
  matcher: space-separated AND terms, each optionally negated or anchored.
  - `foo` — bare term (matching governed by `search_strategy`, see below)
  - `'foo` — toggles the term to the *other* strategy (exact substring in fuzzy
    mode; fuzzy in substring mode)
  - `^foo` — prefix (starts-with)
  - `foo$` — suffix (ends-with)
  - `!foo` — negation (applies to any of the above forms)
  - `foo bar` — AND (both terms must match)
  - `foo | bar` — OR within an AND group (`a b | c` = `a AND (b OR c)`)
- Add a top-level `search_strategy` config field (`fuzzy` | `substring`) that
  chooses what a **bare** term means. Ships defaulting to `fuzzy` (current
  behavior); users who prefer literal matching set `substring`. The `'` prefix
  always toggles to the other strategy, mirroring fzf's `--exact`.
- The extended matcher applies uniformly at every navigation level (providers,
  owners, repos) and in the flat "all repositories" view.
- Update the filter placeholder to hint the syntax.

No breaking change: the shipped default (`fuzzy`, bare terms) reproduces
today's behavior for anyone who doesn't opt in; the new operators are additive.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `tui`: the "Fuzzy filter" requirement becomes an extended-search matcher
  (AND/OR, negation, prefix/suffix/exact anchors) whose bare-term behavior is
  driven by configuration; level-scoping and facet composition are unchanged.
- `configuration`: a new `search_strategy` field with its default and
  validation.

## Impact

- `internal/tui/fuzzy.go` — grows from one subsequence function into a query
  compiler (`compileQuery`) plus per-token matchers; the subsequence primitive
  is retained as the fuzzy kind.
- `internal/tui/model.go` — the three call sites (`filtered[T]` for providers +
  owners, and the inline repo match in `visibleRepos`) route through the
  compiled matcher; the placeholder text is updated.
- `internal/config/config.go` — new `SearchStrategy` field, default, validation.
- `internal/tui/*_test.go`, `internal/config/*_test.go` — new unit tests for the
  operators, strategy flip, and config default/validation.
- No new third-party dependency (matcher is hand-written over `strings`).
