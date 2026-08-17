## 1. Sort state and logic

- [ ] 1.1 Add `sortKey` (`none | name | updated`) and `sortDir` (`asc | desc`)
      types and fields to `Model` in `internal/tui/model.go`; zero value is
      `none`/`asc` (fetch order)
- [ ] 1.2 Add a `sortRepos([]repoItem)` helper using `sort.SliceStable` — name
      compares case-insensitively, updated compares `UpdatedAt`; `asc`/`desc`
      flips the result; `none` returns the slice unchanged
- [ ] 1.3 Call the sort step at the end of `visibleRepos()`, after the filter
      step, so it orders exactly the visible subset at every scope

## 2. Keybindings

- [ ] 2.1 Add `s` in `internal/tui/update.go` to advance `sortKey` through
      `none → name → updated → none`
- [ ] 2.2 Add `S` to toggle `sortDir` (no-op when `sortKey == none`)
- [ ] 2.3 Reset cursor/offset sensibly after a re-sort (clamp cursor to the new
      row count)

## 3. Footer

- [ ] 3.1 Extend `footerText` in `internal/tui/view.go` with `s: sort  S:
      reverse` and show the active key + direction when a sort is active

## 4. Tests

- [ ] 4.1 Test name sort asc/desc ordering (case-insensitive)
- [ ] 4.2 Test last-updated sort asc/desc ordering
- [ ] 4.3 Test `s` cycles `none → name → updated → none` and `none` restores
      fetch order
- [ ] 4.4 Test `S` toggles direction and is a no-op under `none`
- [ ] 4.5 Test sort composes with an active fuzzy filter (only visible repos
      are ordered)

## 5. Verify

- [ ] 5.1 `gofmt -l .` returns empty
- [ ] 5.2 `nix flake check` passes (build + tests + coverage gate)
