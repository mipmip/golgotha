## 1. Config: search_strategy field

- [x] 1.1 Add `SearchStrategy string `yaml:"search_strategy"`` to the top-level
      `Config` struct in `internal/config/config.go`, with a typed constant or
      alias for `fuzzy`/`substring`
- [x] 1.2 Default `SearchStrategy` to `"fuzzy"` in the defaults pass
- [x] 1.3 Validate `SearchStrategy ∈ {fuzzy, substring}` with an actionable
      error naming the bad value and the allowed set
- [x] 1.4 Unit tests: default is `fuzzy` when omitted, `substring` accepted,
      unknown value rejected with a clear message

## 2. Matcher: fzf-style extended search

- [x] 2.1 In `internal/tui/fuzzy.go` add query types: `token{kind, negate,
      text}` with `kind ∈ {fuzzy, substring, prefix, suffix}`, `orGroup`,
      `matcher`
- [x] 2.2 Retain the subsequence loop as `subseqMatch`; implement per-token
      matching (`substring`→`Contains`, `prefix`→`HasPrefix`, `suffix`→
      `HasSuffix`), all case-insensitive
- [x] 2.3 Implement `compileQuery(query string, strategy) matcher`: split on
      whitespace into AND terms, split each on `|` into OR alternatives, parse
      leading `!`, `'`, `^` and trailing `$`; bare term uses `strategy`, `'`
      toggles to the opposite
- [x] 2.4 Implement `matcher.match(target) bool` (lowercase target once, empty
      query matches everything)
- [x] 2.5 Unit tests for every operator, the strategy flip in both modes,
      negation, AND, OR-within-AND, empty query, and mixed expressions

## 3. Wire the matcher into the TUI

- [x] 3.1 Add a `searchStrategy` field to the `Model` and set it from config at
      construction (implemented as `Model.searchMatcher`, reading `cfg`)
- [x] 3.2 Change `filtered[T]` to accept a compiled `matcher` (or query +
      strategy) instead of calling `fuzzyMatch`; update `visibleProviders` and
      `visibleOwners` call sites
- [x] 3.3 Replace the inline `fuzzyMatch` in `visibleRepos` with the compiled
      matcher, preserving the flat-view provider-short participation
- [x] 3.4 Update the filter placeholder to hint the syntax (e.g.
      `"filter — ' exact  ^prefix  $suffix  !not"`)
- [x] 3.5 TUI tests: substring vs fuzzy default at the repo level, `'` toggle,
      prefix/suffix/negation, and level-scoped filtering still works

## 4. Docs & verification

- [x] 4.1 Note `search_strategy` and the filter syntax in BRIEFING.md /
      config reference and the example `config.yaml`
- [x] 4.2 `gofmt -l .` empty; `nix flake check` passes (build + tests +
      coverage gate: core ≥80%, overall ≥70%)
