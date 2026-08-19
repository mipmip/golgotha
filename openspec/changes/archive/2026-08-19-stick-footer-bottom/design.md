## Context

`View()` (internal/tui/view.go) builds the screen by concatenation: rendered
header elements, a blank separator, the windowed body, then the footer elements
— with no vertical fill. The body is windowed to `visible = m.height - chrome()`
rows, so a long list fills the space and the footer sits at the bottom. But a
short body renders at its actual (small) height and the footer floats up beneath
it, leaving empty space at the bottom.

## Goals / Non-Goals

**Goals:**

- Pin the footer block to the bottom of the viewport for short lists.
- No change when the list already fills the viewport or when height is unknown.

**Non-Goals:**

- Changing body alignment (body stays top, directly under the header).
- Changing windowing, scrolling, or `chrome()`.
- Any config/schema change.

## Decisions

### Decision: Pad between body and footer to reach `m.height`

Compute the rendered heights and insert blank lines before the footer block:

```
used = headerHeight + (1 blank separator if header non-empty)
     + bodyHeight
     + footerHeight
pad  = m.height - used
if m.height > 0 && pad > 0 { insert `pad` newlines before the footer }
```

Heights use `lipgloss.Height` so multi-line bodies (fetch progress, "(no
repositories)") are counted correctly.

- **Why self-correcting:** when the list is long, the windowed body height equals
  `visible = m.height - chrome()`, so `used == m.height` and `pad == 0` — the
  footer is already at the bottom and nothing changes. Padding only applies to
  short lists.
- **Why guard on `m.height > 0`:** before the first `WindowSizeMsg` (and in unit
  tests) the height is unknown; padding would be meaningless, so it is skipped.

### Decision: Pin the whole footer block, not just the action menu

All configured footer elements (filter, facet_status, status_message,
position_indicator, action_menu) move to the bottom together; the gap sits
between the body and the first footer element. This matches the mode's footer
being a single block.

- **Alternative — pin only `action_menu`, keep indicator/status hugging the
  body:** more complex (two padded regions) and less predictable. Rejected.

## Risks / Trade-offs

- **[Off-by-one / trailing newline]** → padding math must not push content past
  the last row. Mitigation: derive `pad` from `lipgloss.Height` of the assembled
  parts and clamp to `>= 0`; add a test asserting the rendered height equals
  `m.height`.
- **[Body alignment expectations]** → users might expect the body centered.
  Accepted: footer-at-bottom with body-at-top is the requested behavior.

## Migration Plan

1. Restructure `View()` to build header/body/footer blocks, compute `pad`, and
   insert blank lines before the footer.
2. Add a TUI test: short list → rendered height == `m.height`, footer on the last
   line; full list → unchanged.
3. `gofmt` clean; `go test ./...` and `nix flake check` pass.

## Open Questions

- None.
