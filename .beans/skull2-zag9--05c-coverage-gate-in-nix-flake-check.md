---
# skull2-zag9
title: 05c Coverage gate in nix flake check
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:39:17Z
parent: skull2-ecwd
---

Wire the coverage threshold into `nix flake check` so builds fail below target.

## Tasks
- [x] Coverage measured in test run
- [x] Overall >=70% enforced; core >=80%
- [x] Failure below threshold fails `nix flake check`
- [x] Documented in CLAUDE.md

## Summary of Changes

scripts/coverage.sh gate (overall >=70, core >=80, exit non-zero on fail) wired into flake.nix checks.coverage; nix flake check enforces it. Overall 81.2%. Documented in CLAUDE.md.
