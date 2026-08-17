## Why

`skull2` is a throwaway codename. The project is graduating to a real
identity: **Golgotha** — Aramaic for "the place of the skull" — which
captures the tool's core image of one local vault holding every repo
("brain") harvested from every forge. The command is `gol`: short, unique,
and derived from the brand. Renaming now, while the project is still
PoC/alpha with no external users, avoids a painful rename after adoption.

## What Changes

- **BREAKING** Go module path `github.com/mipmip/skull2` →
  `github.com/mipmip/golgotha`.
- **BREAKING** CLI binary and entrypoint `cmd/skull2` (`skull2 <subcommand>`)
  → `cmd/gol` (`gol <subcommand>`).
- **BREAKING** Auth env-var prefix `SKULL2_<PROVIDER>_TOKEN` →
  `GOLGOTHA_<PROVIDER>_TOKEN`.
- **BREAKING** Config directory `~/.config/skull2/` → `~/.config/golgotha/`
  and cache directory `~/.cache/skull2/` → `~/.cache/golgotha/` (XDG-honored).
- Nix packaging: `subPackages = [ "cmd/gol" ]`, package/binary name `gol`,
  `nix run . -- version` continues to work against the renamed entrypoint.
- Coverage gate (`scripts/coverage.sh`) core-package import paths updated to
  the new module path.
- Docs rebranded: `README.md`, `BRIEFING.md`, `CLAUDE.md`,
  `config.example.yaml`, `openspec/config.yaml` context.
- GitHub remote renamed `mipmip/skull2` → `mipmip/golgotha` (done on GitHub;
  local `jj`/git remote URL updated to match).

## Capabilities

### New Capabilities
<!-- none: this change introduces no new behavior -->

### Modified Capabilities
- `configuration`: config-directory path and the auth env-var prefix
  referenced by the spec change from `skull2`/`SKULL2_` to
  `golgotha`/`GOLGOTHA_`.
- `provider-clients`: token env-var names change from `SKULL2_GITHUB_TOKEN`,
  `SKULL2_CODEBERG_TOKEN`, `SKULL2_GITLAB_TOKEN` to the `GOLGOTHA_` prefix.

## Impact

- **Code**: every Go file's import paths (module rename), `cmd/skull2/`
  directory rename, config-path and cache-path literals in
  `internal/config` and `internal/cache`, env-var literals in
  `internal/provider` and all `*_test.go` fixtures, `e2e/e2e_test.go`.
- **Packaging**: `flake.nix` (`subPackages`, package name), `go.mod`.
- **Tooling**: `scripts/coverage.sh` import paths, `.beans.yml`.
- **Docs**: `README.md`, `BRIEFING.md`, `CLAUDE.md`, `config.example.yaml`,
  `openspec/config.yaml`.
- **External**: GitHub repo slug (`git@github.com:mipmip/golgotha.git`) —
  breaks existing clone URLs; acceptable at alpha with no downstream users.
- **Users**: no data migration shipped; existing local `~/.config/skull2`
  and `~/.cache/skull2` are not auto-moved (documented as a manual step, if
  any pre-alpha user exists).
