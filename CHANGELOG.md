# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Release process: `VERSION` single source of truth (embedded via `go:embed`,
  read by the flake and overridden by goreleaser from the git tag).
- goreleaser config building linux/amd64+arm64 and darwin/amd64+arm64 with
  tar.gz archives and checksums.
- GitHub Actions release workflow triggered on `v*` tag pushes.
- `scripts/release.sh`: gated, interactive release (nix flake check gate,
  vendorHash auto-update, jj-first commit + git tag push).
- `RELEASING.md` maintainer documentation.
- goreleaser and gum in the flake devShell.
