---
# skull2-rw2c
title: release-management
status: todo
type: task
priority: normal
created_at: 2026-08-17T19:05:30Z
updated_at: 2026-08-17T19:36:53Z
parent: skull2-qati
---

see this change and apply this to this project

## OpenSpec change

Captured as `add-release-process` (openspec/changes/add-release-process/) — adapts the jjay/teejay pattern (ref: gh.speclib/jjay .../2026-06-02-release-process): VERSION file + go:embed + flake readFile, goreleaser (4 platforms), GH Actions on v* tags, gum release.sh with nix-flake-check gate + vendorHash auto-update + jj-first hybrid tag/push, CHANGELOG.md + RELEASING.md. New caps version-embedding + release-automation. Validated. Ship with: `/ship add-release-process`.
