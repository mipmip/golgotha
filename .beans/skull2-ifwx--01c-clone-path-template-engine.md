---
# skull2-ifwx
title: 01c Clone-path template engine
status: completed
type: epic
priority: high
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:06:02Z
parent: skull2-j0od
---

Render clone target paths from the configurable `clone_pattern_tpl` Go text/template with the documented field set and per-provider override.

## Tasks
- [x] Renderer with fields: BaseDir, Provider, Type, Short, Host, Owner, OwnerLower, Repo, RepoLower
- [x] Global default + per-provider override resolution
- [x] `~`/BaseDir expansion, cleaned absolute paths, traversal safety
- [x] Unit tests incl. github TechNative-B-V/foo -> ~/gh.technative-b-v/foo

## Summary of Changes

Implemented `internal/clonepath`: text/template renderer (missingkey=error) with all 9 fields, per-provider override, base_dir/`~` expansion, absolute cleaning and traversal rejection. Default renders github TechNative-B-V/foo -> ~/gh.technative-b-v/foo. Coverage 88%.
