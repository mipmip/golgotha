## 1. Normalize the example

- [x] 1.1 Rewrite commented keys in `config.example.yaml` to copy-pasteable `# key: value` form (e.g. `api_url`), so every schema key appears as a `key:` line; no behavioral change (still parses + validates)

## 2. Example-gate test (levels a + b)

- [x] 2.1 Add a go.mod walk-up helper (climb parents from CWD until `go.mod`, return repo root) in an `internal/config` test file
- [x] 2.2 Level (a): test loads `<root>/config.example.yaml` via `config.LoadFrom()` and asserts no error (validation + strict-decoding/unknown-key coverage)
- [x] 2.3 Level (b): reflect over the config struct graph (`Config`/`Provider`/`Auth` and any nested structs), dereferencing pointers and descending into slice/struct elements, collecting each `yaml` tag (strip `,omitempty` etc.)
- [x] 2.4 Level (b): assert each collected key appears as a line-anchored `^\s*#?\s*<key>:` match in the example text; fail listing any missing keys
- [x] 2.5 Add an explicit `allowlist` set (empty initially) of keys intentionally omitted, excluded from the completeness assertion

## 3. Verification

- [x] 3.1 `go test ./internal/config/...` passes; `gofmt -l .` is empty
- [x] 3.2 `nix flake check` passes (the gate rides the existing `coverage` check — confirm no `flake.nix` change was needed)
- [x] 3.3 Keep the `skull2-cqi8` beans checklist current as tasks complete
