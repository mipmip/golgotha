## Why

With configuration in place, the remaining foundations are the clone-path
template engine (used by sync and the TUI to decide where a repo lands) and the
provider/auth abstraction that every concrete provider client will implement.
Both are prerequisites for milestone 02.

## What Changes

- Add a clone-path template engine rendering the configurable
  `clone_pattern_tpl` Go text/template with the documented field set
  (BRIEFING.md section 5), including per-provider override, `~`/BaseDir
  expansion, and path-traversal safety.
- Add the provider abstraction: a `Repo` domain model, a `Provider` interface
  for listing repositories, an auth resolver (configured CLI token → env PAT →
  clear error), and a registry/factory keyed by provider type.

## Capabilities

### New Capabilities
- `clone-path-template`: rendering clone target paths from the configurable
  template and its data fields.
- `provider-abstraction`: the Provider interface, Repo model, and auth
  resolution shared by all concrete provider clients.

### Modified Capabilities
<!-- none -->

## Impact

- `internal/clonepath` and `internal/provider` (new implementations).
- No new third-party dependencies expected (standard library only).
