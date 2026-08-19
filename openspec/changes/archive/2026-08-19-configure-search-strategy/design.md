## Context

The TUI filter (`internal/tui/fuzzy.go`) is a hand-written subsequence matcher:
`fuzzyMatch(target, query)` returns true when every rune of `query` appears in
`target` in order. It is called from three sites in `internal/tui/model.go`:
`filtered[T]` (providers and owners levels, line ~488) and inline in
`visibleRepos` (line ~556, matching `owner/name`, plus the provider short in the
flat view). Subsequence matching produces many surprising hits (`nivis` matches
`nix-services`), and there is no way to ask for literal matching.

We are adopting fzf's extended-search grammar, a widely known convention, so the
default a bare term uses is configurable and every anchor/negation operator is
available per-term.

## Goals / Non-Goals

**Goals:**

- Case-insensitive substring matching reachable both by config default and by
  the `'` per-term toggle.
- fzf-style operators: `'` exact/fuzzy toggle, `^` prefix, `$` suffix, `!`
  negation, space = AND, `|` = OR within an AND group.
- One compiled matcher shared by all three call sites and all navigation levels.
- Backward compatible: shipped default reproduces today's behavior.
- No new third-party dependency; pure `strings`-based implementation.

**Non-Goals:**

- Regex matching, scoring/ranking of results (matcher stays boolean pass/fail).
- Matching against description/README or other metadata (scope stays names/keys,
  as today).
- A visible toggle key or footer mode indicator (config + `'` prefix cover the
  need; can be revisited later).
- Persisting a per-session strategy override.

## Decisions

### Decision: fzf extended-search grammar with a configurable default term type

A query is split on whitespace into AND terms. Each term may contain `|` to form
an OR sub-group. Each atom is parsed for a leading `!` (negation), a leading `'`
(strategy toggle) or `^` (prefix), and a trailing `$` (suffix). A bare atom uses
the configured `search_strategy`.

`search_strategy` maps onto fzf's `--exact`: it decides what a bare term means,
and `'` always flips to the other. This is the least surprising mapping for
anyone who knows fzf, and it makes "substring by default" a one-line config
change.

*Alternatives considered:* a dedicated toggle key cycling modes (more UI, hidden
state, and still no per-term anchors); a separate `exact:`/`fuzzy:` config with
no operators (less powerful, non-standard). fzf's grammar subsumes both.

### Decision: Compile the query once per render into a matcher value

Introduce `compileQuery(query string, strategy SearchStrategy) matcher` in
`fuzzy.go`, where `matcher` holds `[]andGroup`, each `andGroup` a slice of `or`
alternatives, each alternative a `token{kind, negate, text}` with
`kind ∈ {fuzzy, substring, prefix, suffix}`. `matcher.match(target) bool`
lowercases `target` once and evaluates all groups. The existing subsequence loop
is retained as the `fuzzy` kind (`subseqMatch`); `substring/prefix/suffix` use
`strings.Contains/HasPrefix/HasSuffix`.

Compiling once per render (not per item) keeps parsing off the hot path. The
Model gains a `searchStrategy` field set from config at construction; the three
call sites take a compiled matcher instead of calling `fuzzyMatch` directly:
`filtered[T]` accepts a `matcher`, and `visibleRepos` builds one from
`m.filter.Value()` + `m.searchStrategy`.

### Decision: Config field, default, and validation

Add `SearchStrategy string `yaml:"search_strategy"`` to the top-level `Config`
(`internal/config/config.go`). Default it to `"fuzzy"` in the defaults pass;
validate it is one of `{"fuzzy", "substring"}` with an actionable error, mirroring
existing validation style. A small typed constant/alias keeps the TUI and config
in sync.

### Decision: Placeholder hint

Update the filter placeholder (`model.go` ~line 279) from
`"fuzzy filter (owner/name)"` to a concise syntax hint, e.g.
`"filter — ' exact  ^prefix  $suffix  !not"`, so the operators are discoverable
without docs.

## Risks / Trade-offs

- [`$` and `'` are legal characters in repo/owner names] → They are rare in
  practice and fzf has the same ambiguity; anchors are only interpreted at term
  boundaries (leading `^`/`'`/`!`, trailing `$`). Document the behavior; a
  literal `$` mid-term is not special.
- [OR (`|`) parsing is the fiddliest part of the grammar] → It is isolated to
  `compileQuery` and fully unit-tested; if it ever needs to be cut, dropping `|`
  leaves AND/negation/anchors intact. Kept in scope because tests are cheap.
- [Behavior change for existing users who typed subsequence queries] → The
  shipped default stays `fuzzy`, so bare-term behavior is unchanged; only
  operator characters (`^ $ ! ' |`) gain new meaning, which were previously
  matched literally and rarely typed.
- [flat-view target contains a space (`"gh mipmip/huphop"`)] → With space-AND
  semantics, `gh mipmip` becomes two AND terms matched against the whole target;
  both substrings are present, so the flat-view filter keeps working (and becomes
  more predictable than the old subsequence-over-a-space match).
