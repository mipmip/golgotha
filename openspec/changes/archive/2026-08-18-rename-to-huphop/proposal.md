## Why

The project is settling on its real identity: **HupHop**, command **`hup`**
(short, unique, greppable — unlike `hh`, which collides with the `hstr` shell-
history alias). This supersedes the interim `golgotha`/`gol` name. Renaming now,
still at PoC/alpha with no external users, is cheap; the GitHub repo has already
been renamed to `mipmip/huphop`.

## What Changes

- **BREAKING** Go module path `github.com/mipmip/golgotha` →
  `github.com/mipmip/huphop`.
- **BREAKING** CLI binary/entrypoint `cmd/gol` (`gol <cmd>`) → `cmd/hup`
  (`hup <cmd>`).
- **BREAKING** Auth env-var prefix `GOLGOTHA_<PROVIDER>_TOKEN` →
  `HUPHOP_<PROVIDER>_TOKEN`.
- **BREAKING** Config dir `~/.config/golgotha/` → `~/.config/huphop/` and cache
  dir `~/.cache/golgotha/` → `~/.cache/huphop/` (XDG-honored).
- Packaging/tooling: `flake.nix` (`pname`/`mainProgram`/`subPackages = ["cmd/hup"]`),
  `.goreleaser.yaml` (binary `hup`, main `./cmd/hup`), `.github/workflows/release.yml`,
  `scripts/release.sh`, `scripts/coverage.sh` core import paths, root `version.go`
  package name.
- Docs/config rebranded: `README.md`, `BRIEFING.md`, `CLAUDE.md`, `RELEASING.md`,
  `CHANGELOG.md`, `config.example.yaml`, `openspec/config.yaml` context, and the
  golgotha references in the non-delta main specs.
- Local `jj`/git remote repointed to `git@github.com:mipmip/huphop.git` (GitHub
  repo already renamed).

## Capabilities

### New Capabilities
<!-- none: no new behavior -->

### Modified Capabilities
- `configuration`: config-directory path and auth env-var prefix change to
  `huphop`/`HUPHOP_`.
- `provider-clients`: token env-var names change to the `HUPHOP_` prefix.

## Impact

- Every Go file's imports (module rename), `cmd/gol` → `cmd/hup`, path/env
  literals in `internal/config`, `internal/cache`, `internal/provider` and all
  tests + `e2e`. Packaging (`flake.nix`, `go.mod`, `.goreleaser.yaml`, workflow),
  tooling (`scripts/*.sh`), docs. `vendorHash` should stay valid (module rename
  adds no deps). `.beans.yml` `prefix` intentionally left `skull2-` to keep
  existing bean IDs valid.
