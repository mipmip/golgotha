## Context

HupHop's TUI is a single hardcoded experience: fixed chrome (breadcrumb top,
help/keys bottom, contextual strip above the footer) built by string
concatenation in `view.go`, and a fixed primary action (Enter → detail). Bean
`skull2-wzbf` wants a second mode for use inside tmux: strip the chrome, and make
Enter clone-if-needed then run a user `switch_command`.

Introducing modes generalizes the chrome, which is exactly what the (now retired)
`configurable-tui-chrome` proposed — so that work is absorbed here: management is
just one mode. The clone machinery already exists (`Cloner`/`syncer`,
`repoItem.Target`/`Cloned`); `clonepath.Render` computes local paths but does not
expose the path as a template field.

## Goals / Non-Goals

**Goals:**

- A mode selected at launch (`--mode` > `default_mode` > built-in management).
- Per-mode chrome via ordered header/footer element slots (mode-centric).
- Multiplex primary action: clone-if-needed → render → execute `switch_command`.
- Shell-safe command execution; friendly one-line config.
- Backward compatible: no `modes:` ⇒ built-in management mode.

**Non-Goals:**

- In-app mode switching (mode fixed at launch; a later enhancement).
- Self-account tint (rides `fix-self-owner-resolution`).
- Columns (`skull2-n3i2`), star sort (`skull2-2h8p`), combined view
  (`add-combined-repo-view`).
- Generating tmux configuration (explicit BRIEFING non-goal — we run a command).

## Decisions

### Decision: Mode-centric config, absorbing configurable-tui-chrome

Config gains `default_mode` and `modes: { <name>: { header, footer, <settings> } }`.
Each mode carries ordered `header`/`footer` element-name lists (vocabulary:
`breadcrumb`, `action_menu`, `filter`, `facet_status`, `status_message`,
`position_indicator`, plus multiplex-dedicated `switch_hint`, `clone_status`).
The repo list is the implicit body.

- **Why mode-centric over element-centric** ("elements tag their modes"): explicit
  per-mode ordered lists avoid cross-mode ordering ambiguity and make a mode's
  chrome obvious at a glance; "more modes later" is just more blocks.
- **Why absorb rkyi**: the mode-aware system generalizes the flat `tui:` model
  (management = one mode). Shipping flat then generalizing would migrate the
  config twice.
- **Back-compat**: omitting `modes:` yields a built-in management mode equal to
  today's layout.

### Decision: Chrome as an element registry, rendered for the active mode

Refactor `view.go` to a registry `map[string]func(*Model) string`; header/footer
render by iterating the active mode's name lists (skipping empty results); the
body renders between them. `chrome()` counts rendered (wrap-aware) lines so
`viewport.go` windowing stays correct. (This is the retired change's design,
now parameterized by the active mode.)

### Decision: Mode = chrome preset + primary action + settings

A `mode` bundles its element lists, a primary-action behavior, and its settings.
`management`'s action is the current Enter→detail. `multiplex`'s action is
clone-if-needed → `switch_command`. The action is a small strategy keyed by mode,
not a sprawl of `if mode == ...` across the update loop.

### Decision: switch_command — template, then shell-words split, then exec

Render the `switch_command` template against the repo with a context of the
clone-path fields **plus `Target`** (from `clonepath.Render`). Split the rendered
string into argv using POSIX shell-quoting rules (a small splitter) and exec
directly — **no shell** — so names with spaces/metacharacters are literal args.

- **Why**: user picked friendly one-line config + shell safety. Split-not-shell
  gives both. A power-user escape (explicit shell form) can be added if a command
  truly needs pipes/`&&`.
- **Alternative — `sh -c`**: injection risk from repo/owner names. Rejected as
  default.
- **Alternative — argv list in config**: safe but verbose; loses the one-liner.
  Rejected as default; may be offered as an alternate form.

### Decision: multiplex Enter flow

```
activate repo → cloned? ─no→ clone via Cloner (progress) ─fail→ report, stop
                   │yes                    │ok
                   └──────────┬────────────┘
                              → render switch_command (+Target) → exec → report result
```

Reuses the existing `Cloner`. Only render→split→exec→result is new muscle.

## Risks / Trade-offs

- **[Config break]** → `modes:` is a new model. Mitigation: built-in management
  default when omitted; update example + `BRIEFING`; the config-example gate
  (`skull2-cqi8`) keeps the example valid.
- **[Command execution surface]** → running user commands with templated repo
  data. Mitigation: no-shell argv exec; the template only injects repo metadata;
  document the trust model (it is the user's own config).
- **[Shell-words edge cases]** → quoting rules are fiddly. Mitigation: a
  well-tested splitter with explicit cases; offer the argv-list form as a fallback.
- **[Large change]** → chrome system + mode framework + multiplex in one.
  Mitigation: internal seam between the chrome system and command execution;
  land after `fix-self-owner-resolution` to avoid config-file churn.
- **[`Target` needs a resolvable path]** → depends on clone-path config being
  valid. Mitigation: reuse `clonepath.Render` and surface its errors.

## Migration Plan

1. Land `fix-self-owner-resolution` (shared config files) first.
2. Add `default_mode`/`modes` (built-in management default), the element registry
   rendering, and the `--mode` flag; today's behavior becomes `management`.
3. Add multiplex mode: primary action, `Target` in the template context,
   render→shell-words→exec, clone-before-switch.
4. Update `config.example.yaml` (gate-enforced), `BRIEFING.md`,
   `openspec/config.yaml`.

## Open Questions

- After a successful `switch_command`, does the TUI **stay** (jump again) or
  **quit**? tmux `switch-client` leaves the hup pane behind. Resolve during
  implementation; default leaning: stay, with a configurable/keyed quit.
- Offer an explicit argv-list form of `switch_command` in addition to the
  one-line string? Deferred unless a real command needs a shell.
