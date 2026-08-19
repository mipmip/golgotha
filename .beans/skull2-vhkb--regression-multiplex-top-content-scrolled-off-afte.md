---
# skull2-vhkb
title: 'Regression: multiplex top content scrolled off after footer-pin'
status: completed
type: bug
created_at: 2026-08-19T21:30:21Z
updated_at: 2026-08-19T21:30:21Z
parent: skull2-ok4c
---

The stick-footer-bottom change (commit 8053bf66) padded the view to exactly
m.height lines but still appended a trailing newline after the footer, so the
frame occupied m.height+1 line positions. The alt-screen scrolled up by one,
pushing the top row off-screen — most visible in multiplex/--flatlist where the
top row is the first repo.

## Fix

View() now assembles the frame as a []string of lines and joins them WITHOUT a
trailing newline, padding the pre-footer lines up to `m.height - footerLines` so
the footer is pinned to the bottom and the frame is exactly m.height lines. This
also fixes a latent off-by-one for full lists. Footer tests updated to count
lines as Count("\n")+1 and to assert the top line holds content. Hotfix on main
(no OpenSpec change; the "footer pinned to bottom" spec is unchanged).
