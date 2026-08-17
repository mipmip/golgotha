---
# skull2-rw2c
title: release-management
status: completed
type: task
priority: normal
created_at: 2026-08-17T19:05:30Z
updated_at: 2026-08-17T21:04:37Z
parent: skull2-qati
---

see this change and apply this to this project

## OpenSpec change

Captured as `add-release-process` (openspec/changes/add-release-process/) — adapts the jjay/teejay pattern (ref: gh.speclib/jjay .../2026-06-02-release-process): VERSION file + go:embed + flake readFile, goreleaser (4 platforms), GH Actions on v* tags, gum release.sh with nix-flake-check gate + vendorHash auto-update + jj-first hybrid tag/push, CHANGELOG.md + RELEASING.md. New caps version-embedding + release-automation. Validated. Ship with: `/ship add-release-process`.

## Summary of Changes

Shipped via add-release-process (commit 2f23bb37). VERSION file single source of truth (go:embed via root package, ldflags override preserved; flake reads it via builtins.readFile). goreleaser (4 platforms + checksums), GitHub Actions release.yml on v* tags, gated scripts/release.sh (nix flake check + gum bump + CHANGELOG promotion + vendorHash auto-update + jj-first commit and git tag push), CHANGELOG.md + RELEASING.md, goreleaser+gum in devShell. Overall coverage 80.5%.
