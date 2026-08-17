## Context

skull2 is a Go CLI built via a plain nix flake, versioned by jj (colocated git),
shipped to GitHub (mipmip/skull2). Version is hardcoded in `flake.nix:18` and
`cmd/skull2/main.go:23` — it will drift. This adapts jjay's release pattern
(itself from teejay): VERSION file + goreleaser + an interactive release script
that also updates the nix vendorHash.

## Goals / Non-Goals

**Goals:**
- Single source of truth for the version (VERSION file).
- Automated multi-platform releases via goreleaser + GitHub Actions.
- An interactive, gated release script with nix vendorHash auto-update.
- Maintainer docs.

**Non-Goals:**
- Package-manager distribution (homebrew, nixpkgs), binary signing, Docker
  images, automated changelog generation from commits.

## Decisions

### VERSION file + go:embed (ADR, folded in)

Plain-text `VERSION` (e.g. `0.1.0`, no `v` prefix) at the repo root. The binary
reads it via `//go:embed VERSION`; the flake reads it via
`builtins.readFile ./VERSION`; goreleaser overrides via ldflags from the git
tag. One file, two consumers, tag as the release-time override.

_Alternatives: ldflags-only (flake can't read ldflags → still needs a source);
git-tag-as-source (no version for dev builds). Rejected — VERSION avoids drift
and gives dev builds a real version._

### goreleaser for release builds

Keep goreleaser (faithful to the reference) for cross-compilation, checksums and
GitHub-release creation, even though the flake already cross-builds — goreleaser
owns the *release artifacts*, the flake owns *reproducible dev/build*. Targets:
linux/amd64, linux/arm64, darwin/amd64, darwin/arm64. No Windows.

### jj-first hybrid tag/push

skull2 commits via jj (`/ship`, `ship-change.sh`), but GitHub Actions triggers
on a pushed git **tag**, and `jj git push` pushes bookmarks, not arbitrary tags.
So the release script:

```
jj commit  (VERSION + CHANGELOG + vendorHash bump)
jj bookmark set main -r @- ; jj git push --bookmark main
git tag vX.Y.Z @- ; git push origin vX.Y.Z      # tag triggers the workflow
```

### Gate = nix flake check

The pre-release safety check is `nix flake check` (build + tests + coverage
gate), replacing jjay's `make test/build/lint`. The release aborts if it fails.

### Release script with gum

`scripts/release.sh` flow: safety checks (clean tree, on main, CHANGELOG has an
`Unreleased` section, tag absent, `nix flake check` passes) → gum prompt for
bump (major/minor/patch) → update VERSION → promote CHANGELOG `Unreleased` →
version+date → compute + update flake `vendorHash` (fake-hash → `nix build` →
parse expected → write; skip with a warning if nix absent) → jj-first commit +
git tag push. gum absent → print install hint and exit.

### Semantic versioning, `v` prefix on tags

Tags `vX.Y.Z`; VERSION file `X.Y.Z` (no prefix). goreleaser expects the `v`.

## Risks / Trade-offs

- [flake.nix version lags between releases] → the release script updates it; it
  shows the last release version meanwhile. Acceptable.
- [goreleaser duplicates the flake's cross-build] → accepted; different jobs
  (release artifacts vs reproducible build).
- [jj/git tag interplay] → the hybrid push is explicit; the git tag is what
  triggers Actions.
- [gum / nix not installed] → script checks and degrades (hint / skip vendorHash
  with warning).
