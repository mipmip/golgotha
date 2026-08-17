---
# skull2-bror
title: 03a Sync engine
status: todo
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-4uk6
---

Clone-missing + fast-forward-pull-existing across configured owners with dirty-tree safety and a structured per-provider summary.

## Tasks
- [ ] Resolve target path per repo via template engine
- [ ] Clone when absent (respect clone_protocol)
- [ ] fetch + ff-only pull when present; never force
- [ ] Detect dirty trees -> skip + warn
- [ ] Summary struct: cloned/updated/skipped/failed per provider
- [ ] Unit tests with temp git repos
