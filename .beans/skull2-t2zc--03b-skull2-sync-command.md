---
# skull2-t2zc
title: 03b `skull2 sync` command
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:22:47Z
parent: skull2-4uk6
---

Cron-friendly CLI wrapping the sync engine.

## Tasks
- [x] `skull2 sync [--provider NAME] [--no-refresh]`
- [x] Refresh cache first unless --no-refresh
- [x] Line-oriented logging; non-zero exit on any failure
- [x] Integration test of the command path

## Summary of Changes

cmd/skull2 sync [--provider NAME] [--no-refresh]: refresh-first unless --no-refresh, line-oriented logs, non-zero exit on failures; shared selectProviders/refreshProviders helpers.
