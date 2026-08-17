## Why

Every later capability (provider clients, cache, sync, TUI) depends on a
validated configuration that is the single source of truth. Skull2 needs to
load and validate `~/.config/skull2/config.yaml` into typed values with sane
defaults and actionable errors before any of that work can begin.

## What Changes

- Add typed configuration structs for the global settings and per-provider
  settings described in BRIEFING.md section 6 (base_dir, clone_pattern_tpl,
  providers with type/short/urls/clone_protocol/auth/owners/filters).
- Load `~/.config/skull2/config.yaml`, expand `~` in `base_dir`, and apply
  defaults: `base_dir=~`, `clone_protocol=ssh`, `include_archived=false`,
  `include_forks=true`, and the default `clone_pattern_tpl`.
- Validate the config: unique provider names, known provider `type`
  (github/codeberg/gitlab), required fields present, at least one provider,
  and report the first actionable error with file context.
- Add the `skull2 config path` and `skull2 config check` subcommands.

## Capabilities

### New Capabilities
- `configuration`: loading, defaulting and validation of the skull2 YAML
  configuration, plus the `config` CLI subcommands that expose it.

### Modified Capabilities
<!-- none: no existing specs yet -->

## Impact

- `internal/config` (new implementation).
- `cmd/skull2` (new `config` subcommand wiring).
- Adds a YAML dependency; `flake.nix` `vendorHash` must be updated accordingly.
