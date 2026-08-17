## Why

Interactive browsing and cloning is the primary human-facing use case: quickly
find any repository across providers and clone it into the uniform layout, or
open it in the browser.

## What Changes

- Add a Bubble Tea TUI that reads the cache and navigates provider → owner →
  repos, with a global fuzzy filter.
- Single-select clone to the templated target and multi-select bulk clone with
  progress.
- Open the selected repo's `web_url` in the browser; refresh the current
  provider's cache.
- Make the TUI the default command (`skull2` / `skull2 tui`).

## Capabilities

### New Capabilities
- `tui`: the Bubble Tea browse/clone/open interface.

### Modified Capabilities
<!-- none -->

## Impact

- `internal/tui` (new implementation).
- `cmd/skull2` (default to the TUI).
- Adds Bubble Tea dependencies (`bubbletea`, `bubbles`, `lipgloss`); `flake.nix`
  `vendorHash` must be updated.
