---
# skull2-0s31
title: 01b Configuration loading & validation
status: todo
type: epic
priority: high
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-j0od
---

Load and validate `~/.config/skull2/config.yaml` into typed structs with sensible defaults and friendly errors; expose `skull2 config path|check`.

## Tasks
- [ ] Config structs (base_dir, clone_pattern_tpl, providers[], auth, owners, flags)
- [ ] YAML loader with `~` expansion and defaults (base_dir=~, ssh, include_archived=false, include_forks=true)
- [ ] Validation: unique provider names, known types, required fields, actionable errors
- [ ] `skull2 config path` and `skull2 config check`
- [ ] Unit tests incl. a documented example config
