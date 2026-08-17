## 1. Unit coverage

- [ ] 1.1 Fill coverage gaps in config, clonepath, provider, cache, syncer to >=80% each

## 2. End-to-end tests

- [ ] 2.1 E2E harness: httptest provider servers + local bare git repos under t.TempDir()
- [ ] 2.2 E2E: refresh → select → clone lands at the templated path
- [ ] 2.3 E2E: sync clone-missing, then fast-forward after an upstream commit

## 3. Coverage gate

- [ ] 3.1 Coverage script: overall >=70%, core packages >=80%, non-zero on failure
- [ ] 3.2 Add `checks.coverage` to flake.nix wired to the script
- [ ] 3.3 Document the gate in CLAUDE.md

## 4. Verify

- [ ] 4.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...` clean
- [ ] 4.2 `nix flake check` passes with the coverage gate enforced
