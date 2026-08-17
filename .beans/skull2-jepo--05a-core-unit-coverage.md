---
# skull2-jepo
title: 05a Core unit coverage
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:39:17Z
parent: skull2-ecwd
---

Raise unit coverage on core-logic packages (config, template, cache, sync, provider clients) to >=80%.

## Tasks
- [x] Table-driven tests for config/template/cache
- [x] Provider clients tested against mocked HTTP
- [x] Sync engine tested against temp git repos
- [x] >=80% on core packages

## Summary of Changes

Topped up core unit coverage: syncer 75->98%, cache 80->84%. All core packages comfortably >=80% (config 91, clonepath 88, provider 89).
