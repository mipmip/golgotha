---
# skull2-cqi8
title: 'Gate: validate config.example.yaml against the schema'
status: completed
type: task
priority: normal
created_at: 2026-08-19T13:29:26Z
updated_at: 2026-08-19T18:41:43Z
parent: skull2-ok4c
---

Add a quality gate so the documented example config (`config.example.yaml`) can
never drift out of sync with the config schema. From pim's side-note on the
TUI-chrome bean (`skull2-rkyi`): "when changing the config.yaml model, always
update the example config — make this a gate."

## Two levels (both required)

- **(a) Valid** — load `config.example.yaml` through `config.LoadFrom()` +
  `Validate()`; fail if it errors. Catches required-field breakage (e.g. the
  mandatory `username` from `fix-self-owner-resolution`), removed fields the
  example still uses, and — via `Parse()`'s `KnownFields(true)` — typo'd/unknown
  keys for free.
- **(b) Complete** — reflect over `Config`/`Provider`/`Auth`/`TUIConfig`, collect
  every `yaml` tag, and assert each appears as a `key:` (commented `# key:`
  counts) in the example text. This is the level that actually enforces the
  rule: optional-field additions stay *valid* under (a) but (b) catches an
  undocumented new field (e.g. the `tui:` block from `skull2-rkyi`).

Level (a) catches breakage; level (b) catches drift — pim's rule is about drift,
so (b) is the meaningful deliverable, not a stretch.

## Mechanism (decided in explore)

- A **Go test in `internal/config`** doing both (a) and (b). It rides the
  existing `go test ./...` inside the `coverage` flake check — **no flake.nix
  wiring needed** (`src = self` already ships `config.example.yaml` into the
  sandbox). Level (b) must be Go anyway (needs struct reflection).
- Locate the example via **go.mod walk-up**: climb parents from the test's CWD
  until `go.mod`, then read `<root>/config.example.yaml`. (Symlink-in-testdata
  and `go:embed` both rejected — embed can't reach up out of the package dir;
  symlinks are fragile in the nix `src`.)
- Matching for (b): line-anchored `^\s*#?\s*<key>:` to avoid substring false
  positives. Keep a small **allowlist** (empty initially) for keys deliberately
  omitted from the example.

## Implementation note / wrinkle

- Today's example documents `api_url` as prose ("api_url defaults to ...") not as
  `# api_url:` — so building the gate includes **normalizing commented keys to
  `# key: value` form**. Bonus: that makes the example copy-pasteable. Once
  normalized, the gate passes against the current schema, so this can land
  **first**.

## Sequencing / relationships

- Independent; **land first** so it CI-enforces the example updates in
  `fix-self-owner-resolution` (task 5.2, `username`) and `configurable-tui-chrome`
  (task 3.1, `tui:` block).
- Capability: `quality-gate` (spec delta), mirroring the existing coverage gate.
- Related: `skull2-0s31` (config loading/validation framework already exists),
  `skull2-rkyi`, `fix-self-owner-resolution`.

## Optional, separate (NOT this bean)

- `hup config check <path>` as a user-facing feature (validate an arbitrary file
  before installing). Only gives level (a); the gate doesn't need it. File
  separately if wanted.



## Summary of Changes

Added a two-level config-example gate riding the existing coverage check (no
flake wiring): (a) `TestExampleConfigValid` loads config.example.yaml via
`config.LoadFrom`; (b) `TestExampleConfigComplete` reflects over the Config
struct graph and asserts every yaml key appears as a line-anchored `key:`
(commented/list forms count), with an empty allowlist. Normalized the example's
`api_url` comment to `# key:` form. Shipped as 2026-08-19-add-config-example-gate
(commit 25645768).
