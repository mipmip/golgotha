---
# skull2-0s31
title: 01b Configuration loading & validation
status: completed
type: epic
priority: high
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:06:02Z
parent: skull2-j0od
---

Load and validate `~/.config/skull2/config.yaml` into typed structs with sensible defaults and friendly errors; expose `skull2 config path|check`.

## Tasks
- [x] Config structs (base_dir, clone_pattern_tpl, providers[], auth, owners, flags)
- [x] YAML loader with `~` expansion and defaults (base_dir=~, ssh, include_archived=false, include_forks=true)
- [x] Validation: unique provider names, known types, required fields, actionable errors
- [x] `skull2 config path` and `skull2 config check`
- [x] Unit tests incl. a documented example config

## OpenSpec change

Proposed as `add-config-loading` (openspec/changes/add-config-loading/) — proposal, design, specs and tasks authored and validated. Ready for `/opsx:apply`.

## Summary of Changes

Implemented `internal/config`: typed Config/Provider/Auth, XDG/`~/.config` path resolution, strict YAML (KnownFields), defaults + `~` expansion, and Validate(). Wired `skull2 config path|check`. Coverage 91%.
