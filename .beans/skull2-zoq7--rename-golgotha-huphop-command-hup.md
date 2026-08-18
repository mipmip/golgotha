---
# skull2-zoq7
title: Rename golgotha -> HupHop (command hup)
status: in-progress
type: task
priority: normal
created_at: 2026-08-18T08:27:31Z
updated_at: 2026-08-18T08:28:08Z
parent: skull2-qati
---

Third and final rename: module github.com/mipmip/huphop, cmd/hup, HUPHOP_ env, ~/.config|.cache/huphop. Repo already renamed on GitHub.

## OpenSpec change

Captured as `rename-to-huphop`. Full rename golgotha->huphop / gol->hup (module, cmd dir, env prefix, config/cache dirs, flake, goreleaser, workflow, release.sh, coverage.sh, docs, non-delta specs) + configuration/provider-clients spec deltas. Repo already mipmip/huphop; ship repoints the local remote first. Ship with: `/ship rename-to-huphop`.
