---
# skull2-9ftr
title: multiplex mode refinements
status: completed
type: task
priority: normal
created_at: 2026-08-19T19:47:17Z
updated_at: 2026-08-19T20:46:41Z
parent: skull2-ok4c
---

- the checkbox should be hidden. Multiselect is not relevant in this mode
- after switch the application should exit 0
- when cloning needs to be done first a popup with a progress bar should be shown
- extra argument to open the browse in flat listing at startup (one level deeper) e.g. --flatlist



## Design (explore 2026-08-19)

One change `refine-multiplex-mode` covering all four. Builds on shipped
add-tui-modes + add-combined-repo-view.

**1. Hide checkbox / disable multiselect in multiplex.** Mode-specific (NOT the
flat view): in the combined view under *management*, bulk-clone stays useful, so
hide only when the active mode is multiplex (`activeSwitchCommand() != ""`).
renderRepos omits the `[ ]`/`[x]` column; space becomes a no-op in multiplex.

**2. Exit 0 after a successful switch.** Resolves the earlier stay-vs-quit open
question → **quit**. After runSwitch succeeds: set quitting, return tea.Quit
(clean exit 0), ideal for a tmux display-popup launch. Clone/switch *failure*
stays open with an error (no non-zero exit plumbing for now).

**3. Clone-first → centered modal popup with a determinate bar.** Two parts:
  - **Async clone**: multiplexActivate becomes async (a tea.Cmd) so the popup
    can animate; flow: Enter → (uncloned) start clone, show popup → progress msgs
    update bar → done → runSwitch → quit; error → close popup + status; Esc →
    cancel clone, stay. Already-cloned → switch → quit immediately.
  - **Determinate bar (parse git)**: teach the syncer a progress-emitting clone
    that runs `git clone --progress` and parses stderr ("Receiving objects: N%",
    split on \r) into percent+phase, mirroring the fetch.Emit pattern. TUI streams
    these into the popup bar.
  - **Centered modal overlay**: a bordered lipgloss box centered over the dimmed
    list (manual overlay compositing — splice box lines into the background at the
    center offset), not a full-screen replace.

**4. `--flatlist` startup flag.** `hup tui --flatlist` starts one level deeper in
the combined flat view (flatAll=true, nav=levelRepos); caches already load in
New. Orthogonal to --mode. **Killer combo:** `hup tui --mode multiplex --flatlist`
= flat fuzzy list of every repo everywhere → Enter → clone-if-needed → switch →
exit 0 (the whole tmux quick-jump in one command).

**Touches:** internal/syncer (progress clone), internal/tui (multiplex.go async +
popup overlay, view.go renderRepos, update.go space/Enter), cmd/hup + run.go
(--flatlist). Config schema unchanged (no gate/example impact).



## Summary of Changes

Four multiplex refinements. (1) renderRepos omits the checkbox and toggleSelect
is inert when multiplexActive(). (2) A successful switch sets quitting + returns
tea.Quit (exit 0); failures stay with an error. (3) Syncer gained a
progress-emitting clone: ExecGit.CloneProgress runs `git clone --progress`,
scanCRorLF splits stderr, parseGitCloneProgress → (frac,phase); Engine.CloneRepoProgress
emits via callback (plain CloneRepo unchanged for sync). (4) multiplexActivate is
async: streams a cloneEvent channel, shows a centered lipgloss.Place modal with a
determinate bubbles/progress bar, Esc cancels; on done → switch → quit. (5)
`hup tui --flatlist` starts in the combined flat view via startFlat; composes
with --mode. Coverage overall 80.5% (syncer 89.4%). Shipped as
2026-08-19-refine-multiplex-mode (commit 372c99d0).
