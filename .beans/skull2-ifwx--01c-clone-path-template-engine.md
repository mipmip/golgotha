---
# skull2-ifwx
title: 01c Clone-path template engine
status: todo
type: epic
priority: high
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-j0od
---

Render clone target paths from the configurable `clone_pattern_tpl` Go text/template with the documented field set and per-provider override.

## Tasks
- [ ] Renderer with fields: BaseDir, Provider, Type, Short, Host, Owner, OwnerLower, Repo, RepoLower
- [ ] Global default + per-provider override resolution
- [ ] `~`/BaseDir expansion, cleaned absolute paths, traversal safety
- [ ] Unit tests incl. github TechNative-B-V/foo -> ~/gh.technative-b-v/foo
