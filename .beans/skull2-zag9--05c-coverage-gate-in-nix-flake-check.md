---
# skull2-zag9
title: 05c Coverage gate in nix flake check
status: todo
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-ecwd
---

Wire the coverage threshold into `nix flake check` so builds fail below target.

## Tasks
- [ ] Coverage measured in test run
- [ ] Overall >=70% enforced; core >=80%
- [ ] Failure below threshold fails `nix flake check`
- [ ] Documented in CLAUDE.md
