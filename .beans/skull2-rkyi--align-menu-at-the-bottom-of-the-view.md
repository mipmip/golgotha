---
# skull2-rkyi
title: Configurable TUI header/footer elements
status: scrapped
type: feature
priority: normal
created_at: 2026-08-17T21:45:51Z
updated_at: 2026-08-19T18:24:42Z
parent: skull2-ok4c
---

Make the TUI chrome configurable: `header` and `footer` are ordered slots of
placeable elements; the repo list is always the fixed body. Supersedes the
original "align menu at the bottom" idea — you place `action_menu` in the
footer (the default) or the header instead of a single top/bottom flag.

## Config shape

```yaml
tui:
  header: [breadcrumb]
  footer: [filter, facet_status, status_message, position_indicator, action_menu]
```

These defaults reproduce today's layout exactly, so a config without a `tui:`
block behaves identically (backward compatible). Empty lists render no chrome
(e.g. `header: []` / `footer: []`).

## Element vocabulary (all placeable)

| element              | source today            | notes                          |
|----------------------|-------------------------|--------------------------------|
| `breadcrumb`         | `headerText()`          | always-on                      |
| `action_menu`        | `footerText()` (keys)   | always-on; may wrap → multiline|
| `filter`             | `filter.View()`         | conditional + interactive      |
| `facet_status`       | `facets.status()`       | conditional                    |
| `status_message`     | `m.status`              | conditional                    |
| `position_indicator` | `indicatorText`         | contextual (X–Y of N)          |

Body (repo list, `bodyText()`) is implicit — always the middle, top-aligned
(unchanged). Never listed as an element.

## Behavior / rules

- Order within a list = render order (top-to-bottom).
- Validation: element names must be known; each element appears at most once
  across header+footer combined; empty lists are legal.
- Conditional elements render only when active; `chrome()` counts rendered
  non-empty lines dynamically so windowing (`viewport.go`) stays correct.
- `chrome()` must count *rendered* lines (action_menu can wrap on narrow
  terminals) — fixes a latent fixed-height assumption.

## Implementation sketch

Refactor `internal/tui/view.go` from hardcoded concatenation into an element
registry: `map[name]func(m) string`. Header/footer render by iterating their
configured name lists (skipping empty results). Body window logic unchanged.
Add `TUIConfig{ Header, Footer []string }` under `Config.TUI` in
`internal/config`, with defaults + validation.

## Dependencies / relationships

- **Changes the config model** → must update `config.example.yaml`, and is a
  prime beneficiary of the config-example validation gate (`skull2-cqi8`).
- **Foundation for `skull2-wzbf` (tui-modes):** per-mode presets layer on top —
  multiplex mode = empty header/footer ("no menu"). Mode wiring is refined in
  that bean, not here.
- Sequencing note: both this and `fix-self-owner-resolution` touch the config
  schema + example; land after (or coordinate with) that change to avoid
  example/schema churn.

## Open (decide at propose time)

- Whether `filter` in the header (interactive input at top) needs any focus
  affordance, or is acceptable as-is.



---
**Scrapped 2026-08-19:** folded into `skull2-wzbf` (add-tui-modes). The mode-aware
chrome supersedes this standalone flat `tui:` model — management is just one mode.
The `configurable-tui-chrome` OpenSpec change was retired.
