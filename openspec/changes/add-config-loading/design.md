## Context

Skull2 has takeoff scaffolding (`internal/config` is an empty package). This
change implements configuration loading — the first real capability and the
dependency of every other milestone. The schema is fixed by BRIEFING.md
section 6; this design covers how to load, default, and validate it in Go.

## Goals / Non-Goals

**Goals:**
- Typed, validated config as the single source of truth.
- Documented defaults and `~` expansion.
- Actionable errors suitable for cron logs.
- `skull2 config path|check`.

**Non-Goals:**
- Rendering clone paths (epic 01c, `internal/clonepath`).
- Resolving auth tokens or talking to providers (epic 01d / milestone 02).
- Writing or migrating config files.

## Decisions

- **YAML library**: use `gopkg.in/yaml.v3`. It is the de-facto standard, supports
  strict decoding, and integrates cleanly with struct tags. Alternative
  (`sigs.k8s.io/yaml`) adds JSON round-tripping we do not need.
- **Strict decoding**: enable `KnownFields(true)` so typos in keys fail loudly
  rather than being silently ignored.
- **Defaults before validation**: unmarshal, then fill defaults, then validate,
  so validation sees the effective config.
- **Path resolution**: resolve the config path via `$XDG_CONFIG_HOME` when set,
  else `~/.config/skull2/config.yaml`; expand `~` using `os.UserHomeDir`.
- **Error style**: return wrapped errors (`fmt.Errorf("%w")`) with the field and
  provider name; no panics; first error wins for `config check` output.
- **Validation lives in the package**: a `Config.Validate() error` method keeps
  rules next to the types and unit-testable without the CLI.

## Risks / Trade-offs

- [Strict decoding rejects forward-compatible keys] → acceptable for a PoC;
  documented so future keys are added to the structs deliberately.
- [Adding the YAML dep changes `vendorHash`] → update `flake.nix` `vendorHash`
  in the same change; `nix flake check` will surface the expected hash.
