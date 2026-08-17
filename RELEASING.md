# Releasing golgotha

This document describes how maintainers cut a release. Releases are automated:
pushing a `vX.Y.Z` git tag triggers a GitHub Actions workflow that runs
goreleaser to build multi-platform binaries and publish a GitHub release.

## How versioning works

The version lives in **one** place — the `VERSION` file at the repo root
(e.g. `0.1.0`, no `v` prefix). Three consumers read it:

- The Go binary embeds it via `go:embed` (`version.go` → `golgotha.Version`,
  used as the default of `main.version` in `cmd/gol/main.go`).
- `flake.nix` reads it via `builtins.readFile ./VERSION`.
- goreleaser overrides `main.version` from the **git tag** via ldflags at
  release time (the tag wins over the embedded default).

## Prerequisites

- `jj` and `git` (the repo is jj colocated with git).
- `gum` — interactive prompts. Provided by the flake devShell (`nix develop`)
  or install standalone.
- `nix` — runs the pre-release gate (`nix flake check`) and recomputes the
  flake `vendorHash`. If absent, the script warns and skips those steps.
- `goreleaser` — only needed for local validation; the release itself runs it
  in CI. Provided by the devShell.
- Push access to `git@github.com:mipmip/golgotha.git`.

Enter the dev shell to get gum + goreleaser on PATH:

```bash
nix develop
```

## Pre-release checklist

- [ ] Working tree is clean (`jj diff` shows nothing).
- [ ] You are on `main`.
- [ ] `CHANGELOG.md` has an `## [Unreleased]` section describing the changes.
- [ ] The target `vX.Y.Z` tag does not already exist.
- [ ] `nix flake check` passes (build + tests + coverage gate).

The release script enforces all of these and aborts before making any change if
one fails.

## Cutting a release

```bash
./scripts/release.sh
```

The script:

1. Runs the safety checks above (aborts on any failure).
2. Prompts (via gum) for the bump: `patch`, `minor`, or `major`; computes the
   new version and the `vX.Y.Z` tag.
3. Updates `VERSION`.
4. Promotes the CHANGELOG `## [Unreleased]` section to `## [X.Y.Z] - YYYY-MM-DD`
   and leaves a fresh empty `Unreleased` above it.
5. Recomputes and writes the flake `vendorHash` (fake-hash → `nix build` →
   parse expected → write). Skipped with a warning if nix is absent.
6. Commits the bump with jj, sets the `main` bookmark, and `jj git push`.
7. Creates the `vX.Y.Z` git tag and pushes it — this triggers the release
   workflow.

## Verifying the release

- Watch the workflow: <https://github.com/mipmip/golgotha/actions>.
- Confirm the GitHub release contains the four archives
  (`golgotha_X.Y.Z_{linux,darwin}_{amd64,arm64}.tar.gz`) and `checksums.txt`:
  <https://github.com/mipmip/golgotha/releases>.
- Optionally, verify a downloaded binary reports the right version:

  ```bash
  tar -xzf golgotha_X.Y.Z_linux_amd64.tar.gz
  ./gol version   # → gol vX.Y.Z
  ```

## Local validation (optional)

Before releasing you can validate the goreleaser config and a snapshot build
without pushing anything:

```bash
goreleaser check
goreleaser build --snapshot --clean   # builds all four targets into dist/
```

## Troubleshooting

- **"working tree is not clean"** — commit or discard changes; `jj diff` must
  be empty.
- **"not on 'main'"** — the `main` bookmark must point at `@` or `@-`. Set it
  with `jj bookmark set main -r @`.
- **"CHANGELOG.md has no 'Unreleased' section"** — add a `## [Unreleased]`
  heading with the pending changes.
- **"tag vX.Y.Z already exists"** — that version was already released; pick a
  different bump.
- **`nix flake check` fails** — fix the failing build/test/coverage before
  releasing. The gate exists to keep broken releases out.
- **vendorHash update skipped / mismatch** — if nix is absent the script keeps
  the existing hash. It only changes when Go dependencies change; if CI's nix
  build later fails on the hash, run the script in an environment with nix, or
  update `vendorHash` manually (build with a fake hash and copy the `got:`
  value nix reports).
- **Workflow did not trigger** — it triggers only on `v*` tags. Confirm the tag
  was pushed: `git push origin vX.Y.Z`. `jj git push` pushes bookmarks, not
  arbitrary tags, which is why the script pushes the tag with plain git.
- **Release job fails on `GITHUB_TOKEN`** — the workflow uses the default
  token with `contents: write`; ensure Actions are enabled for the repo.
