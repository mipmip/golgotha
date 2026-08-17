## Why

skull2 needs a repeatable release process before its first release. Version is
hardcoded in two places (`flake.nix` and `cmd/skull2/main.go`), which will
drift, and there is no automation for building release binaries or cutting
GitHub releases. This adapts the proven jjay/teejay release pattern
(goreleaser + release script + nix vendorHash update) to skull2's jj + nix
toolchain.

## What Changes

- Add a `VERSION` file as the single source of truth for the version.
- Embed the version via `go:embed` in `cmd/skull2/main.go` (keep the ldflags
  override for goreleaser).
- Read the version in `flake.nix` via `builtins.readFile ./VERSION` (replacing
  the hardcoded `0.0.0-dev`).
- Add `.goreleaser.yaml` building linux/amd64, linux/arm64, darwin/amd64,
  darwin/arm64 with checksums, version injected from the git tag.
- Add `.github/workflows/release.yml` triggered on `v*` tag pushes, running
  goreleaser to create the GitHub release.
- Add `scripts/release.sh` (gum): safety checks (incl. `nix flake check`),
  interactive bump, update VERSION + CHANGELOG + flake vendorHash, then a
  jj-first commit/push with a git tag push to trigger Actions.
- Add `CHANGELOG.md` (Keep a Changelog, with an `Unreleased` section) and
  `RELEASING.md` maintainer docs.
- Add goreleaser and gum to the flake devShell.

## Capabilities

### New Capabilities
- `version-embedding`: single `VERSION` file, `go:embed` in the binary,
  `builtins.readFile` in the flake, ldflags override precedence.
- `release-automation`: goreleaser config, GitHub Actions release workflow, and
  the gated `scripts/release.sh` (nix flake check + vendorHash auto-update +
  jj-first tag/push).

### Modified Capabilities
<!-- none: flake.nix is scaffolding, not a spec'd capability; changes are impact -->

## Impact

- New files: `VERSION`, `.goreleaser.yaml`, `.github/workflows/release.yml`,
  `scripts/release.sh`, `CHANGELOG.md`, `RELEASING.md`.
- Modified: `cmd/skull2/main.go` (go:embed), `flake.nix` (readFile version;
  goreleaser + gum in devShell).
- Dev dependencies (devShell only): goreleaser, gum. Releases are triggered by
  pushing a `v*` tag.
- Independent of the queued feature changes (fetch-progress / repo-filters /
  repo-details); can ship in any order relative to them.
