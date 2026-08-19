## Context

`config.example.yaml` (repo root) is the canonical documentation of the config
schema. It is not loaded by any test today, so schema changes drift it silently.
The config loader already exists: `config.LoadFrom(path)` → `Parse` (strict
`KnownFields(true)`) → `applyDefaults` → `Validate`. The `coverage` flake check
runs `scripts/coverage.sh`, which runs `go test ./...` from the repo root with
`src = self`, so the example ships into the sandbox.

Two drift modes must be caught: **invalid** (required field added, field removed,
key typo) and **stale-but-valid** (optional field added, left undocumented). The
first is caught by loading; the second only by comparing the schema's field set
against the example text.

## Goals / Non-Goals

**Goals:**

- Fail `nix flake check` when `config.example.yaml` is invalid (level a) or
  incomplete (level b).
- Zero new CI wiring: ride the existing `go test ./...`.
- Keep the example the single source of truth (no duplicated fixture).

**Non-Goals:**

- Any schema/behavior change.
- A `hup config check <path>` CLI feature (separate, optional).
- Validating value *semantics* beyond what `Validate()` already checks.

## Decisions

### Decision: Go test in `internal/config`, riding the coverage check

Both levels live in a `_test.go` in `internal/config`. Level (b) needs struct
reflection over the config types, so Go is the only sensible home; level (a) is
a plain `LoadFrom` call. Running under `go test ./...` means the existing
`coverage` check enforces it with no `flake.nix` change.

- **Alternative — a dedicated `checks.config-example` flake output** running the
  CLI or a script: more visible as a named gate, but duplicates what `go test`
  already runs and can't do level (b) without the structs. Rejected.

### Decision: Locate the example via go.mod walk-up

Go runs tests with CWD set to the package dir (`internal/config`), while the
example is at repo root. A helper climbs parent directories until it finds
`go.mod`, then reads `<root>/config.example.yaml`.

- **Alternative — `go:embed`**: cannot embed files above the package directory,
  and the example is two levels up. Rejected.
- **Alternative — a `testdata` symlink** to the root file: idiomatic Go, keeps a
  single source, but symlinks in the nix `src` (`self`) are a fragility risk.
  Rejected in favor of the dependency-free walk-up.

### Decision: Reflection-based completeness with line-anchored matching

Level (b) walks the config struct graph (`Config` → `Provider` → `Auth`, plus
future nested structs like `TUIConfig`), dereferencing pointers and descending
into slice/struct element types, collecting each field's `yaml` tag (stripping
options like `,omitempty`). For each key it asserts a line-anchored match
`^\s*#?\s*<key>:` in the example text, so a commented `# owners:` counts as
documented but prose mentions do not.

- **Allowlist:** a small, explicit set (empty initially) of keys intentionally
  omitted from the example, so omissions stay visible and reviewed.
- **Why line-anchored + colon:** avoids substring false positives (a key name
  appearing inside a comment sentence) and forces the copy-pasteable `key:`
  form.

### Decision: Normalize the example to `# key: value` comment form

Today `api_url` is documented as prose ("api_url defaults to ..."), which the
colon-anchored check would not credit. Part of this change rewrites such
commented keys into `# key: value` form. This is a documentation-quality
improvement (copy-pasteable) and a behavioral no-op — the example already
parses/validates.

## Risks / Trade-offs

- **[False negatives in level (b)]** → a key documented only in prose is flagged.
  Mitigation: normalize the example (this change) and use the allowlist for true
  exceptions.
- **[Walk-up finds the wrong root]** in unusual layouts (nested modules) →
  Mitigation: stop at the first `go.mod`; the repo is a single module, so this is
  unambiguous here.
- **[Reflection misses a nested type]** added later without a yaml tag path →
  Mitigation: recurse generically over structs/pointers/slices so new nested
  config types are covered automatically.
- **[Gate blocks unrelated work]** if someone adds a field and forgets the
  example → that is the intended behavior; the failure message names the missing
  key(s).

## Migration Plan

1. Normalize `config.example.yaml` commented keys to `# key: value`.
2. Add the walk-up helper, the reflection collector, the allowlist, and the two
   assertions.
3. `nix flake check` passes against the current schema, confirming the baseline.
4. Land this change **first**; subsequent schema changes
   (`fix-self-owner-resolution`, `configurable-tui-chrome`) must update the
   example or this gate fails their `nix flake check`.

## Open Questions

- None blocking. The allowlist starts empty; entries are added only with an
  explicit reason if a field should ever be undocumented.
