---
# skull2-zoq7
title: Rename golgotha -> HupHop (command hup)
status: completed
type: task
priority: normal
created_at: 2026-08-18T08:27:31Z
updated_at: 2026-08-18T08:36:22Z
parent: skull2-qati
---

Third and final rename: module github.com/mipmip/huphop, cmd/hup, HUPHOP_ env, ~/.config|.cache/huphop. Repo already renamed on GitHub.

## OpenSpec change

Captured as `rename-to-huphop`. Full rename golgotha->huphop / gol->hup (module, cmd dir, env prefix, config/cache dirs, flake, goreleaser, workflow, release.sh, coverage.sh, docs, non-delta specs) + configuration/provider-clients spec deltas. Repo already mipmip/huphop; ship repoints the local remote first. Ship with: `/ship rename-to-huphop`.

## Summary of Changes

Shipped via rename-to-huphop (commit 56416a8f). Full rebrand golgotha -> HupHop (binary hup): module github.com/mipmip/huphop, cmd/hup, HUPHOP_ env prefix, ~/.config/huphop + ~/.cache/huphop, flake (pname/mainProgram/subPackages cmd/hup), goreleaser/release workflow/release.sh, coverage.sh import paths, root version.go package, docs and config.example, non-delta specs. GitHub repo + local remote already mipmip/huphop. nix flake check green; hup 0.1.0 builds; vendorHash unchanged. Overall coverage 81.3%. .beans.yml prefix left skull2- to preserve existing bean IDs.
