## Why

`config.example.yaml` is the canonical, user-facing documentation of the config
schema, but nothing verifies it stays in sync. When the schema changes the
example silently drifts: it can become **invalid** (a new required field, a
removed field it still uses) or merely **stale** (a new optional field left
undocumented). This is imminent — `fix-self-owner-resolution` adds a mandatory
`username` (invalidates the example) and `configurable-tui-chrome` adds an
optional `tui:` block (would leave it stale-but-valid). Bean `skull2-cqi8`
captures the rule: *"when changing the config model, always update the example —
make this a gate."*

## What Changes

- Add a Go test in `internal/config` that enforces two levels against the repo's
  `config.example.yaml`, riding the existing `go test ./...` run inside the
  `coverage` flake check (no `flake.nix` wiring needed):
  - **(a) Valid**: load the example via `config.LoadFrom()` + `Validate()`; fail
    on error. Also catches unknown/typo'd keys for free (strict decoding uses
    `KnownFields(true)`).
  - **(b) Complete**: reflect over `Config`/`Provider`/`Auth` (and any nested
    config structs), collect every `yaml` tag, and assert each appears as a
    `key:` (a commented `# key:` counts) in the example text.
- Locate the example via **go.mod walk-up** (climb parent dirs from the test's
  working directory until `go.mod`, then read `<root>/config.example.yaml`).
- Matching for (b) is line-anchored (`^\s*#?\s*<key>:`) with a small **allowlist**
  (empty initially) for keys deliberately omitted from the example.
- **Normalize `config.example.yaml`** so every optional/commented key uses the
  copy-pasteable `# key: value` form (today `api_url` is written as prose, which
  the completeness check would not credit). No schema change.

Out of scope: a `hup config check <path>` CLI feature (would only give level (a)
and is not needed for the gate; file separately if wanted).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `quality-gate`: add a requirement that the documented example config is both
  valid against the schema and complete (documents every schema field), enforced
  by `nix flake check`.

## Impact

- **Tests**: `internal/config` — a new example-gate test (levels a + b), plus a
  go.mod walk-up helper and an allowlist. Test-only; no production code paths
  change. Adds coverage to the `config` core package.
- **Docs/example**: `config.example.yaml` normalized to `# key: value` comment
  form (behavioral no-op).
- **CI**: no `flake.nix` change — the test runs under the existing `coverage`
  check (`src = self` already ships the example into the sandbox).
- **Sequencing**: lands **first** of the three in-flight changes; then it
  CI-enforces the example updates in `fix-self-owner-resolution` (task 5.2) and
  `configurable-tui-chrome` (task 3.1).
