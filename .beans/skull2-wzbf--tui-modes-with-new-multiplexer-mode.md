---
# skull2-wzbf
title: tui-modes with new multiplexer mode
status: in-progress
type: feature
priority: normal
created_at: 2026-08-19T11:37:59Z
updated_at: 2026-08-19T19:01:25Z
parent: skull2-ok4c
---

I've made this draft as addition to the configuration file:

```
default_mode: management_mode

# Default mode or start with --mode management
management_mode: 
  menu_position: bottom    

# Start with --mode multiplex
multiplex_mode:
  switch_command: tmux-helper open-or-new-session-with-name="{{.Short}}->{{.OwnerLower}}" open-or-new-window-with-name="{{.Repo}}" layout-template="3-pane-coding"
```

You can understand what i want to achieve:

- a new mode to be used inside tmux to quickly navigate to a session with a window containing a checkout of a repo.
- only relevant interface items should be visiable. no menu (i guest)
- when a repo is not cloned it should be cloned just before the switch_command is executed

- the current main mode is now called management_mode and is the default mode. This default can be overridden.

## Related: combined cross-provider repo view (skull2-39es)

pim wants a combined list of repos across **all providers and organizations**
in one flat view. Now tracked as its own feature bean `skull2-39es` (split off
from the self-owner bug fix `skull2-qp5y`). It fits the "modes" theme — likely a
view/mode concern rather than a new nav level — so keep it in mind when
designing modes here.


## Decisions (explore 2026-08-19)

**Scope:** one change `add-tui-modes` that **supersedes/absorbs
`configurable-tui-chrome` (skull2-rkyi)** — the mode-aware chrome IS the
configurable chrome (management = one mode). Milestone-07 later-stretch; aligns
with BRIEFING goal #2 ("quick navigation"). A runtime `switch_command` is NOT
the "tmux generator" non-goal — we run a user command, not generate config.

**A mode = { chrome preset } + { primary action } + { mode config }.**
- chrome: per-mode header/footer element lists (mode-centric).
- primary action: management Enter → detail; multiplex Enter → clone-if-needed → switch.
- mode config: e.g. multiplex `switch_command`.

**Config shape (mode-centric, chosen):**
```yaml
default_mode: management        # overridable by --mode <name>
modes:
  management:
    header: [breadcrumb]
    footer: [action_menu, position_indicator]
  multiplex:
    header: []
    footer: [switch_hint]       # multiplex-dedicated elements allowed
    switch_command: 'tmux-helper new-session -s "{{.Short}}->{{.OwnerLower}}" -c {{.Target}}'
```
Element vocabulary carries over from rkyi (breadcrumb, action_menu, filter,
facet_status, status_message, position_indicator) plus multiplex-dedicated ones
(switch_hint, clone_status, …). Body (repo list) always fixed. More modes later.

**switch_command execution:** render the friendly one-line string (Go template),
then **shell-words split → argv, executed WITHOUT a shell** (injection-safe for
repo/owner names). Optional explicit `sh -c` form for power users needing pipes.

**Template context gap:** switch_command needs `{{.Target}}` (resolved local
clone path) — the clone template exposes BaseDir/Provider/Type/Short/Host/Owner/
OwnerLower/Repo/RepoLower but NOT the computed path. Add Target to the context.

**Multiplex Enter flow:** cloned? → (no) clone via existing Cloner w/ progress,
fail ⇒ don't switch; → render switch_command (+Target) → exec → handle result.
Open: after switch, TUI stays vs quits (tmux switch-client leaves the pane).

**Mode selection:** `--mode` flag > `default_mode` config; fixed at launch
(multiplex is launched from a tmux keybind). In-app switching optional/later.

**Consequences:**
- `configurable-tui-chrome` change is retired/superseded; `skull2-rkyi` → scrapped
  (folded here).
- This change now TOUCHES config schema ⇒ guarded by the config-example gate
  (skull2-cqi8) and shares config files with `fix-self-owner-resolution`
  (sequence after it). Depends on nothing else.
- Still large for one change; the chrome-system vs command-execution seam is the
  natural internal split if it ever needs breaking up.
