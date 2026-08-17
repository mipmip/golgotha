---
# skull2-ci84
title: 'TUI: lazy-fetched org shows no repos until restart'
status: in-progress
type: bug
priority: high
created_at: 2026-08-17T21:50:46Z
updated_at: 2026-08-17T21:50:46Z
parent: skull2-qati
---

Race: fetch Done emitted before cache save; UI reload sees empty. Fixed by committing cache before emitting Done.
