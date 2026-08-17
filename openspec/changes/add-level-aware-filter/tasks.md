## 1. Level-scoped filtering

- [ ] 1.1 Add a filter helper / per-level visible lists (visibleProviders, visibleOwners) using fuzzyMatch on raw names
- [ ] 1.2 Update `bodyText` to filter the current level's items instead of always `visibleRepos()`

## 2. Enter & lifecycle

- [ ] 2.1 Remove the "flatten to repos" shortcut in `enter()`; drill the highlighted filtered item per level
- [ ] 2.2 Clear the filter on any level change (drill in and back); reset cursor/offset

## 3. Tests & verify

- [ ] 3.1 Update-driven tests: filter narrows providers/owners/repos at each level; Enter drills a filtered owner; match ignores label decorations; filter clears on drill-in and back
- [ ] 3.2 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
