## Why

Browsing is faster when you can inspect a repository before cloning. Pressing
Enter on a repo should open a detail view — description, stars, topics, language,
last updated, and the rendered README — so you can decide without leaving the
TUI or opening a browser.

## What Changes

- Make repo-level **Enter open a detail view** instead of cloning; cloning stays
  on `c` (Enter now consistently "drills deeper" at every level).
- Add a detail view showing tier-1 metadata (already cached: description,
  updated_at, urls, archived, fork, visibility) plus lazily-fetched tier-2
  (stars, topics, language) and tier-3 (README).
- Render the README with **glamour** (in-process markdown), scrollable via
  `bubbles/viewport`. (Supersedes the bean's original `bat` note; glamour avoids
  external-ANSI-in-TUI issues but adds a Go dependency → flake vendorHash bump.)
- Fetch details lazily on first open and cache them **separately** from the list
  cache (per-repo files), storing raw markdown; `r` re-fetches. Render at view
  time so width/theme adapt.
- Degrade gracefully offline: if the fetch fails, show tier-1 metadata and a
  "README unavailable" note — never an error screen.

## Capabilities

### New Capabilities
- `repo-details`: lazy fetch of repository detail (stars/topics/language +
  README), a separate per-repo detail cache, glamour rendering, and graceful
  offline behavior.

### Modified Capabilities
- `provider-abstraction`: add repository detail + README fetch to the provider
  interface.
- `provider-clients`: implement detail + README fetch per provider.
- `tui`: Enter opens the detail view; `c` is the sole clone key.

## Impact

- `internal/provider` (interface + client mappings), `internal/cache` or a new
  detail-cache module, `internal/tui` (detail level + view). Adds the
  `charmbracelet/glamour` dependency (update `flake.nix` vendorHash).
- Sequenced AFTER `add-fetch-progress` and `add-repo-filters`; this change mostly
  ADDS provider methods, so conflict risk is lowest — do it last.
