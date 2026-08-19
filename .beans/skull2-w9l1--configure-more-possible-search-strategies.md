---
# skull2-w9l1
title: configure more possible search strategies
status: completed
type: feature
priority: normal
created_at: 2026-08-19T23:06:03Z
updated_at: 2026-08-19T23:29:56Z
parent: skull2-ok4c
---

Currently the filter search does a very fuzzy search. I personally prefer a simple case insenstive substring search. When I now search for nivis, I still get tons of results. I really just want to see all project that have the substring nivis in org or repo.

It could be configured gloabally or it could be set in a search mode using conventinal starting keys. Lets chat about this.

## Summary of Changes

Replaced the subsequence-only TUI filter with an fzf-style extended-search
matcher and made the bare-term behavior configurable. New top-level
`search_strategy` config key (`fuzzy` default | `substring`) decides what an
unanchored term means; a `'` prefix always toggles a term to the opposite
(mirrors fzf `--exact`). Operators: `'exact`, `^prefix`, `suffix$`, `^…$`
equality, `!negate`, space = AND, `a | b` = OR within an AND group.

Implementation: `compileQuery(query, strategy) → matcher` in
`internal/tui/fuzzy.go` (subsequence primitive kept as `subseqMatch`);
`Model.searchMatcher` compiles once and all three call sites
(`visibleProviders`/`visibleOwners`/`visibleRepos`) route through it, so the
strategy and operators apply at every navigation level and in the flat view.
Config field defaulted + validated in `internal/config/config.go`; documented in
`config.example.yaml`, `BRIEFING.md`, `CHANGELOG.md`. Filter placeholder now
hints the syntax. Tests: config default/accept/reject, `TestCompileQuery` (all
operators + both-mode strategy flip), and repo/providers-level wiring tests.

Verified via `nix flake check` (build + tests + coverage gate: core ≥80%,
overall ≥70%). Shipped as `2026-08-19-configure-search-strategy` (commit
5f042dff).
