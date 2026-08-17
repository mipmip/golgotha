---
# skull2-vle5
title: more filters
status: in-progress
type: task
priority: normal
created_at: 2026-08-17T18:13:11Z
updated_at: 2026-08-17T20:16:15Z
parent: skull2-qati
---

- forked yes/no
- public/private
- archived yes/no

## OpenSpec change

Captured as `add-repo-filters` (openspec/changes/add-repo-filters/) — Model A narrow-only tri-state facets (fork/archived) + visibility value-cycle, Repo.Visibility string mapped across providers; deltas to provider-abstraction, provider-clients and tui. Validated. Ships AFTER `add-fetch-progress`. Ship with: `/ship add-repo-filters`.
