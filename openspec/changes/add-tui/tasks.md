## 1. Dependencies

- [ ] 1.1 Add bubbletea, bubbles, lipgloss; `go mod tidy`; update flake.nix vendorHash

## 2. Model & navigation

- [ ] 2.1 Root model with navigation stack (provider → owner → repos) reading the cache
- [ ] 2.2 Help/footer bar with keybindings; `q` quits; already-cloned indicator

## 3. Filter & actions

- [ ] 3.1 Global fuzzy filter (`/`) over flattened repos
- [ ] 3.2 Single clone and multi-select (space) bulk clone via the syncer engine, with progress
- [ ] 3.3 `o` opens web_url; `r` refreshes current provider cache

## 4. Wiring & tests

- [ ] 4.1 Make TUI the default command (`skull2` / `skull2 tui`)
- [ ] 4.2 Update-function tests for navigation, filter and actions (no TTY)
- [ ] 4.3 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` clean
