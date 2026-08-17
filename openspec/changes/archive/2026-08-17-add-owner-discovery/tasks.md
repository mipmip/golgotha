## 1. Configuration

- [x] 1.1 Add `all_owners bool` and `exclude_owners []string` to the provider config
- [x] 1.2 Owner-set resolver: `{self} ∪ {discovered} ∪ {owners} − {exclude}` (case-insensitive); default off = current behavior
- [x] 1.3 Config + resolver unit tests

## 2. Provider discovery

- [x] 2.1 GitHub `/user/orgs` discovery (paginated)
- [x] 2.2 Codeberg `/user/orgs` discovery (paginated)
- [x] 2.3 GitLab member-groups discovery (paginated)
- [x] 2.4 Zero-discovery warning (likely missing scope)
- [x] 2.5 Discovery tests against mocked HTTP

## 3. Cache v2

- [x] 3.1 Owner index with per-owner fetch state (discovered vs fetched); atomic write
- [x] 3.2 Backward-compatible read of the legacy flat cache
- [x] 3.3 Cache tests (round-trip, unfetched owner, legacy read)

## 4. Sync (eager)

- [x] 4.1 On `all_owners`, discover then fetch every resolved owner; update cache; clone/pull
- [x] 4.2 Respect `exclude_owners`
- [x] 4.3 Engine/command tests

## 5. TUI (lazy)

- [x] 5.1 Owner level sourced from the cached owner index (incl. unfetched)
- [x] 5.2 Fetch-on-entry with loading state; cache the result; `r` re-fetches current owner
- [x] 5.3 Update-driven tests (unfetched entry triggers fetch cmd; cached entry instant)

## 6. Verify

- [x] 6.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
- [x] 6.2 Update config.example.yaml with all_owners / exclude_owners documented
