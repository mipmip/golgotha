## Why

Some users drive their repos with Jujutsu (`jj`) rather than plain git and want
HupHop to clone as jj (bean `skull2-wr68`): globally configurable, overridable
per provider, and per repo. Today every clone is `git clone`.

Cloning **colocated** (`jj git clone --colocate`) keeps a real top-level `.git`,
so the entire existing sync path (repo detection, dirty check, fast-forward
pull, `dirtygit`) keeps working unchanged — only the clone step branches.

## What Changes

- Add a `clone_vcs` setting (`git` | `jj`, default `git`) at the global level and
  as a per-provider override, plus per-repo `vcs_rules` (owner/name globs → vcs,
  first match wins). Resolution: `vcs_rules` → provider `clone_vcs` → global
  `clone_vcs` → `git`.
- When the resolved VCS is `jj`, clone with `jj git clone --colocate` (a real
  `.git` remains, so sync is unaffected); otherwise `git clone` as today.
- Guard `jj` availability: if a jj clone is requested and `jj` is not on `PATH`,
  fail with an actionable error. The default stays `git`.
- The multiplex clone popup shows a determinate bar for jj clones too: the
  progress path spawns `jj git clone` under a pseudo-terminal (jj only emits its
  percentage bar on a TTY and has no `--progress` flag), strips ANSI, and parses
  the bare `NN%` (falling back to a spinner if a chunk does not parse).
  `hup sync` uses a plain (non-PTY) jj clone with no bar.
- Document the new keys in `config.example.yaml` (required by the config-example
  gate).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `configuration`: add `clone_vcs` (global + per-provider) and per-repo
  `vcs_rules`, with resolution and validation (`vcs` ∈ {git, jj}; each `match`
  is a valid glob).
- `sync`: the clone engine can clone via `jj git clone --colocate` when the
  resolved VCS is jj (keeping a colocated `.git`); the progress-emitting clone
  reports jj progress from a pseudo-terminal.

## Impact

- **Config**: `internal/config` — `Config.CloneVCS`, `Provider.CloneVCS`,
  `Provider.VCSRules` (`VCSRule{Match, VCS}`), a pure `CloneVCSFor` resolver, and
  validation. Example config updated.
- **Syncer**: `internal/syncer` — branch the clone on the resolved VCS; add a
  colocated jj clone (plain for sync, PTY-based progress for the TUI popup with
  ANSI stripping + `NN%` parsing); guard `jj` on `PATH`.
- **Dependency**: a pty library (e.g. `creack/pty`) for the jj progress path —
  update `flake.nix` `vendorHash`.
- **TUI**: no new code — the multiplex popup inherits the jj bar via the progress
  clone; the git bar is unchanged.
- **Tests**: `CloneVCSFor` resolution + validation; jj clone command selection;
  jj progress parsing (ANSI-stripped `NN%`); jj-missing error.
