---
# skull2-wr68
title: allow clone as jj
status: in-progress
type: task
priority: normal
created_at: 2026-08-18T09:28:59Z
updated_at: 2026-08-19T21:55:11Z
parent: skull2-ok4c
---

should be globally configurable, yet per repo overridable



## Design (explore + spike, 2026-08-19)

Change name: `add-jj-clone`. Clone via `jj git clone --colocate` when the
resolved VCS is jj; everything else stays git.

**Colocated (decided):** `jj git clone --colocate` keeps a real top-level `.git`,
so the whole sync path (IsRepo/.git, IsDirty=git status, Fetch, FastForward) is
UNCHANGED. Only the clone step branches. Stateless (a colocated repo is just a
git repo to huphop). jj imports git-side changes on next jj use.

**Config surface (global → provider → per-repo rules; first match wins):**
```yaml
clone_vcs: git                 # global default: git | jj (sibling of clone_protocol)
providers:
  - name: gh
    clone_vcs: git             # per-provider override
    vcs_rules:                 # per-repo overrides, first match wins
      - { match: "mipmip/dotfiles", vcs: jj }   # path.Match glob on owner/name
      - { match: "mipmip/*",        vcs: jj }
```
Pure resolver `config.CloneVCSFor(provider, "owner/name") -> git|jj`:
vcs_rules(first glob match) → provider.clone_vcs → clone_vcs → git. Validate
vcs ∈ {git,jj} and that each match compiles as a path.Match glob.

**jj-on-PATH guard:** if a resolved jj clone runs and `jj` is missing, fail with
a clear error. Global default stays git.

**Gate synergy:** new keys (clone_vcs, vcs_rules, match, vcs) MUST be documented
in config.example.yaml or the completeness gate (skull2-cqi8) fails.

## Spike findings — jj git clone progress (jj 0.41.0)

- jj shows a percentage bar ONLY on a TTY. There is NO `--progress` flag to force
  it. With stderr piped, jj emits only LF phase lines ("Fetching into new repo",
  "bookmark: main@origin [new] tracked", "Added N files") — no percentages.
- On a PTY, the bar format is bare ` NN% [bar]` (NO phase label, unlike git's
  "Receiving objects: N%"), \r-updated with ANSI clear/cursor codes (ESC[2K,
  ESC[?25l/h) and colored block chars; runs in ~2 passes (percent resets).

**Decision (post-spike): PTY + parse NN% for a determinate jj bar.**
- The progress-emitting clone (TUI multiplex popup only) spawns
  `jj git clone --colocate --color never <url> <dest>` under a pseudo-terminal
  (adds a pty dependency, e.g. creack/pty → update flake vendorHash), strips ANSI
  CSI sequences, splits on \r, parses `(\d+)%` → determinate bar (generic phase
  label since jj gives none). Spinner fallback if a chunk doesn't parse.
- `hup sync` (headless) uses the plain clone (piped stderr, no PTY, no bar).
- So two jj clone paths: plain (sync) and PTY-progress (popup).

## Scope
- config: clone_vcs (global+provider), VCSRule{match,vcs}+vcs_rules, CloneVCSFor,
  validation, example update.
- syncer: jj colocated clone (plain + PTY-progress with ANSI-strip + NN% parse),
  branch clone on resolved vcs, jj-on-PATH guard; pty dependency + vendorHash.
- tui: multiplex popup inherits it (git bar unchanged; jj bar via PTY path).
