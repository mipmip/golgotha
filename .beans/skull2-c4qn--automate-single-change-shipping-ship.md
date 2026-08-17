---
# skull2-c4qn
title: Automate single-change shipping (/ship)
status: completed
type: feature
priority: normal
created_at: 2026-08-17T17:03:30Z
updated_at: 2026-08-17T17:03:30Z
parent: skull2-qati
---

Single gated command to apply, verify, archive, commit and push one OpenSpec change.

## Summary of Changes

- scripts/ship-change.sh: deterministic tail — stage, gate (`nix flake check`: build + tests + coverage >=70/>=80), `openspec archive`, `jj commit` (Pim Snel, no self-promo), push main; aborts on unchecked tasks or gate failure.
- .claude/commands/ship.md: `/ship <change>` orchestrates open-bean -> openspec-apply -> ship-change.sh -> close-bean.

Turns the previously-manual apply -> verify -> archive -> commit loop into one gated command.
