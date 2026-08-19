---
# skull2-q0ik
title: footer should be sticked at the bottom of the viewport
status: completed
type: bug
priority: normal
created_at: 2026-08-19T19:55:36Z
updated_at: 2026-08-19T21:13:49Z
parent: skull2-ok4c
---


## Summary of Changes

Pinned the TUI footer to the bottom of the viewport. View() now builds the
header/body/footer as separate blocks, counts their lines (textLines helper,
empty=0), and inserts m.height-used blank lines before the footer when the body
is short and the height is known. Long lists (pad 0) and unknown height (no pad)
are unchanged. Also fixed a pre-existing flaky test (TestGitHubFetchOwnerCancel)
that raced cancellation against a successful response — the handler now blocks on
r.Context().Done() so cancellation is deterministic. Shipped as
2026-08-19-stick-footer-bottom (commit 8053bf66).
