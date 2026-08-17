---
# skull2-pt7y
title: 05b End-to-end tests
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:39:17Z
parent: skull2-ecwd
---

Prove the PoC end to end against mocked provider APIs and temp dirs.

## Tasks
- [x] Mock provider HTTP servers + fixture repos
- [x] E2E: refresh -> browse -> clone to correct templated path
- [x] E2E: sync clone-missing then ff-pull-existing
- [x] Runnable headlessly under nix

## Summary of Changes

e2e package: hermetic TestRefreshBrowseClone (httptest GitHub fixture -> cache -> clonepath -> clone from file:// bare, lands at templated path) and TestSyncCloneThenFastForward (clone-missing then ff after upstream commit). No network/SSH.
