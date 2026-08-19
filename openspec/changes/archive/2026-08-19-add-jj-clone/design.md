## Context

Every clone today is `git clone` via `syncer.ExecGit.Clone` /
`CloneProgress`. The whole sync path is git-based: `IsRepo` checks `.git`,
`IsDirty` is `git status`, update is `git fetch` + `git merge --ff-only`. Bean
`skull2-wr68` wants jj clones, globally configurable and per-repo overridable.

A **spike** (jj 0.41.0) established two facts:
- `jj git clone --colocate` produces a real top-level `.git`, so all git-based
  sync operations keep working — only the clone step changes.
- `jj git clone` shows its percentage bar **only on a TTY** and has **no
  `--progress` flag**. Piped, it emits only phase lines (no percentages). On a
  PTY the bar is bare ` NN% [bar]` (no phase label) with `\r`+ANSI, in ~2 passes.

## Goals / Non-Goals

**Goals:**

- Configurable clone VCS: global `clone_vcs`, per-provider override, per-repo
  `vcs_rules`.
- Clone colocated when jj so sync is unchanged and stateless.
- Determinate progress bar for jj clones in the multiplex popup.
- Clear error when `jj` is requested but absent.

**Non-Goals:**

- jj-native (non-colocated) repos or a jj-aware sync/fetch path.
- Changing the git clone/sync behavior.
- Progress bars outside the multiplex popup (`hup sync` shows none).

## Decisions

### Decision: Colocated jj clone; sync stays git

Clone with `jj git clone --colocate` when the resolved VCS is jj. The colocated
`.git` means `IsRepo`/`IsDirty`/`Fetch`/`FastForward` are unchanged and nothing
needs to record which VCS was used (stateless). jj imports git-side changes on
the user's next jj command.

- **Alternative — jj-native (non-colocated):** requires a parallel jj VCS
  abstraction across the syncer and breaks git tooling like `dirtygit`. Rejected.

### Decision: Resolution order (pure, testable)

`config.CloneVCSFor(p *Provider, ownerName string) string`:

```
first matching provider.VCSRules entry (glob on "owner/name")
  → provider.CloneVCS
  → global Config.CloneVCS
  → "git"
```

`VCSRule{ Match string; VCS string }`; `Match` uses `path.Match` semantics
(`*` does not cross `/`, so `owner/*` and `*/dotfiles` work). Validation:
`VCS` ∈ {git, jj}; each `Match` compiles via `path.Match`.

### Decision: Two jj clone paths (plain vs PTY-progress)

- **Plain** (used by `hup sync`): `jj git clone --colocate` with stderr piped;
  no progress needed.
- **PTY-progress** (used by the TUI multiplex popup): spawn
  `jj git clone --colocate --color never <url> <dest>` attached to a
  pseudo-terminal so jj emits its bar; strip ANSI CSI sequences, split on `\r`,
  parse `(\d+)%` into fraction events (generic phase label, since jj emits
  none). If a chunk does not parse, keep going (indeterminate/spinner fallback);
  the clone still returns its terminal result.

Adds a pty dependency (e.g. `creack/pty`); update `flake.nix` `vendorHash`.

- **Why PTY:** jj has no `--progress`; piped, it yields no percentages. A TTY is
  the only way to get a determinate bar (the chosen behavior).

### Decision: Guard `jj` on PATH

Before running a jj clone, check `jj` resolves on `PATH`; if not, fail with an
actionable error. The global default stays `git`, so users without jj are
unaffected unless they opt in.

## Risks / Trade-offs

- **[jj progress format drift]** → the ` NN% [bar]` format (no label, ANSI,
  multi-pass) is jj-version-specific. Mitigation: parse defensively (any `\d+%`
  after ANSI stripping) and fall back to indeterminate; the clone completes
  regardless. Only the popup bar is affected.
- **[pty dependency]** → new third-party dep + `vendorHash` churn, and PTY
  behavior can differ across platforms. Mitigation: PTY is used only in the
  progress path; the plain path (and all of `hup sync`) needs no PTY.
- **[Wrong-VCS config]** → a typo in `vcs_rules` could clone unexpectedly.
  Mitigation: validate values/globs at load; document in the example.
- **[jj + git both required]** → a jj-cloned repo is still synced with git.
  Accepted: jj users have git too.

## Migration Plan

1. Add config: `clone_vcs` (global+provider), `vcs_rules`, `CloneVCSFor`,
   validation; document keys in `config.example.yaml` (gate-enforced).
2. Syncer: branch the clone on the resolved VCS; colocated jj clone (plain +
   PTY-progress with ANSI strip / `NN%` parse); guard `jj` on PATH; add the pty
   dep and update `vendorHash`.
3. Verify: `gofmt`, `go test ./...`, `nix flake check` (coverage gate).

## Open Questions

- None blocking. jj's exact progress bytes were captured in the spike; the
  parser targets `\d+%` after ANSI stripping with a spinner fallback.
