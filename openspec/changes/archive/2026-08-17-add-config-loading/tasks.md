## 1. Types

- [x] 1.1 Define `Config` (base_dir, clone_pattern_tpl, providers) in `internal/config`
- [x] 1.2 Define `Provider` (name, type, short, api_url, web_url, clone_protocol, auth, owners, include_archived, include_forks)
- [x] 1.3 Define `Auth` (cli, env) and a `ProviderType` enum/const set (github, codeberg, gitlab)

## 2. Loading

- [x] 2.1 Resolve config path (`$XDG_CONFIG_HOME` or `~/.config/skull2/config.yaml`)
- [x] 2.2 Add `gopkg.in/yaml.v3`; decode with `KnownFields(true)`; update `flake.nix` vendorHash
- [x] 2.3 Return an actionable error when the file is missing or malformed

## 3. Defaults & expansion

- [x] 3.1 Apply defaults (base_dir=~, clone_protocol=ssh, include_archived=false, include_forks=true, default clone_pattern_tpl)
- [x] 3.2 Expand `~`/`~/` in base_dir to an absolute home path

## 4. Validation

- [x] 4.1 Implement `Config.Validate()`: >=1 provider, unique names, known types, required fields
- [x] 4.2 Return the first error with provider/field context

## 5. CLI

- [x] 5.1 Wire `skull2 config path` (print resolved path, exit 0)
- [x] 5.2 Wire `skull2 config check` (print summary / error, exit 0 or non-zero)

## 6. Tests

- [x] 6.1 Table-driven tests: valid config, missing file, malformed YAML, unknown key
- [x] 6.2 Defaults and `~` expansion tests
- [x] 6.3 Validation tests (no providers, duplicate names, bad type, missing fields)
- [x] 6.4 Ensure `nix flake check` passes with the new dependency
