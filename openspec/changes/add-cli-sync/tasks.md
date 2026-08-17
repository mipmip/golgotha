## 1. Git operations

- [ ] 1.1 `internal/syncer`: git runner interface (clone, fetch, ff-only merge, status, rev-parse)
- [ ] 1.2 Real implementation shelling to the `git` binary

## 2. Engine

- [ ] 2.1 For each repo: resolve target via clonepath; clone if absent
- [ ] 2.2 If present: fetch + ff-only pull on default branch; never force
- [ ] 2.3 Dirty tree (`status --porcelain` non-empty) → skip + warn
- [ ] 2.4 Aggregate per-provider summary (cloned/updated/skipped/failed)

## 3. Command

- [ ] 3.1 `skull2 sync [--provider NAME] [--no-refresh]`; refresh cache unless --no-refresh
- [ ] 3.2 Line-oriented logs; non-zero exit if any repo failed

## 4. Tests & verify

- [ ] 4.1 Engine tests with temp git repos (clone-missing, ff-pull, dirty-skip)
- [ ] 4.2 Command wiring test
- [ ] 4.3 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` clean
