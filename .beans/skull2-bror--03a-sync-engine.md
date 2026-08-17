---
# skull2-bror
title: 03a Sync engine
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:22:47Z
parent: skull2-4uk6
---

Clone-missing + fast-forward-pull-existing across configured owners with dirty-tree safety and a structured per-provider summary.

## Tasks
- [x] Resolve target path per repo via template engine
- [x] Clone when absent (respect clone_protocol)
- [x] fetch + ff-only pull when present; never force
- [x] Detect dirty trees -> skip + warn
- [x] Summary struct: cloned/updated/skipped/failed per provider
- [x] Unit tests with temp git repos

## Summary of Changes

internal/syncer: Git runner interface + ExecGit (clone/fetch/ff-only/status/rev-parse); engine resolves target via clonepath, clones missing, ff-pulls clean, skips dirty with warning, per-provider summary. Coverage 90%.
