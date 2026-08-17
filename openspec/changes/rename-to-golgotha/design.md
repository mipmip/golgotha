## Context

`skull2` is a codename. The rename to **Golgotha** (command `gol`) is a
pure identity change — no behavior changes. It is mechanical but wide: it
touches the Go module path (every import), the `cmd/` entrypoint directory,
auth env-var names, XDG config/cache directory names, Nix packaging, tooling
scripts, docs, and the GitHub remote slug. Because the project is still
PoC/alpha with no external users, this is the cheapest moment to do it.

## Goals / Non-Goals

**Goals:**

- Ship a consistent `Golgotha` brand and `gol` command across code,
  packaging, tests, and docs.
- Keep `nix build`, `nix flake check` (build + tests + coverage gate), and
  `gofmt -l .` green after the rename.
- Leave zero references to `skull2`/`SKULL2` outside archived history.

**Non-Goals:**

- No config/cache data migration tooling. Existing local `~/.config/skull2`
  and `~/.cache/skull2` are not auto-moved (documented manual step at most).
- No behavior, schema, or feature changes.
- No changes to the clone-path template or the `Short` provider tags
  (`gh`/`cb`/`gl`) — those are unrelated to the project name.

## Decisions

**Module path → `github.com/mipmip/golgotha`.**
The Go module path must change to match the renamed remote so `go get` and
import paths stay coherent. Mechanically: edit `go.mod`, then rewrite every
`github.com/mipmip/skull2/...` import. Rationale for a hard rename over an
alias: Go has no module-alias mechanism, and leaving the old path would be
permanently confusing. Alternative considered — keep module path, rename only
binary — rejected: the user chose a full rename including the repo slug, so a
mismatched module path would be worse than the churn.

**Entrypoint `cmd/skull2` → `cmd/gol`.**
Directory rename so the built binary is `gol`. `flake.nix` `subPackages`
updates to `cmd/gol`. `nix run . -- version` keeps working because it targets
`packages.default`, not the directory name.

**Env-var prefix `SKULL2_` → `GOLGOTHA_`.**
Chosen the full brand (`GOLGOTHA_GITHUB_TOKEN`) over the short command form
(`GOL_GITHUB_TOKEN`). Env vars are set rarely (once, in a shell profile or CI
secret) so verbosity costs little, and the full brand is unambiguous and
grep-friendly. The `gol` command stays short for interactive/cron use where
typing frequency actually matters. This split (verbose env, terse command)
mirrors common tooling convention.

**XDG dirs `skull2` → `golgotha`.**
`internal/config` and `internal/cache` compute `~/.config/skull2` and
`~/.cache/skull2` (honoring `$XDG_CONFIG_HOME` / `$XDG_CACHE_HOME`). Rename
the path segment to `golgotha`. No migration shim — alpha, no users.

**Execution as a single atomic rename.**
Do the module/env/dir/path renames as one coherent pass driven by grep, then
run the full gate once. A rename that leaves the tree half-renamed won't
compile, so there is no value in staging it across commits. Prefer scripted
find-and-replace on exact tokens (`github.com/mipmip/skull2`, `SKULL2_`,
`"skull2"`, `cmd/skull2`) with a final `grep -ri skull2` sweep to prove zero
residue.

**Remote rename ordering.**
Rename `mipmip/skull2` → `mipmip/golgotha` on GitHub first (GitHub keeps a
redirect), then update the local `jj`/git remote URL. This keeps pushes
working through the transition.

## Risks / Trade-offs

- [Missed reference — a stray `skull2` literal in a test fixture or doc slips
  through] → Final `grep -rIn 'skull2\|SKULL2\|Skull2'` across the tree
  (excluding `.git` and archived openspec) must return only intended history;
  gate on it before commit.
- [Coverage gate breaks because `scripts/coverage.sh` hardcodes the old module
  path in its core-package list] → Explicit task to update those import paths;
  `nix flake check` will fail loudly if missed.
- [`vendorHash` churn] → No dependencies are added or removed, so the module
  rename alone should not change `vendorHash`; if `nix build` reports a
  mismatch, update it from the printed expected hash.
- [Existing clones point at the old URL] → Acceptable at alpha; GitHub's
  automatic redirect covers the interim. Documented in the proposal impact.
- [Pre-alpha user has data under `~/.config/skull2`] → No auto-migration;
  documented as a manual `mv` if anyone is affected (effectively nobody).

## Migration Plan

1. Rename the GitHub repo to `mipmip/golgotha`; update local remote URL.
2. Rename `cmd/skull2` → `cmd/gol`; edit `go.mod` module path.
3. Scripted replace of `github.com/mipmip/skull2` imports across all `.go`.
4. Replace `SKULL2_` → `GOLGOTHA_` and XDG `skull2` → `golgotha` path
   segments in `internal/config`, `internal/cache`, `internal/provider`, and
   all test fixtures.
5. Update `flake.nix`, `scripts/coverage.sh`, `.beans.yml`,
   `config.example.yaml`.
6. Rebrand docs: `README.md`, `BRIEFING.md`, `CLAUDE.md`,
   `openspec/config.yaml`.
7. Run `gofmt -l .` (empty), `nix flake check` (green), and the residue grep
   (zero hits) as the gate.

Rollback: the change is a single commit; `jj` revert restores the prior name.

## Open Questions

None — repo-slug decision (full rename) resolved with the user before
proposing.
