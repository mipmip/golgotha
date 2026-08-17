---
description: Ship one OpenSpec change end-to-end (apply → gate → archive → commit → push → close bean)
---

Ship the OpenSpec change named in the arguments as a single gated step. One
change per invocation. Argument: the change name (kebab-case). If omitted, run
`openspec list` and use the sole active change, or ask which one.

Do this in order; do not skip the gate:

1. **Announce & open the bean.** State the change. Find the linked bean (look for
   an "OpenSpec change" note referencing it, or `beans list -S "<change>"`) and
   mark it in-progress: `beans update <id> -s in-progress`. If no bean exists,
   continue without one.

2. **Implement.** Invoke the openspec-apply-change skill for this change.
   Implement every task with thorough tests and check each off in `tasks.md`
   (`- [ ]` → `- [x]`). Do not stop until all tasks are complete. If genuinely
   blocked, stop and report — do not ship a partial change.

3. **Ship (gated).** Run:
   `bash scripts/ship-change.sh <change> "<commit subject>"`
   This stages the tree, runs the gate (`nix flake check` — build + tests +
   coverage ≥70% overall / ≥80% core), archives the change, commits as Pim Snel
   with no self-promotion, and pushes `main`. Write a clear one-line commit
   subject describing the change. If the gate fails, fix the code and re-run —
   never bypass it.

4. **Close the bean.** Mark the linked bean(s) completed with a
   `## Summary of Changes` section. If the bean is an epic whose parent
   milestone now has all epics completed, mark the milestone completed too.

5. **Report.** Change archived name, commit id, coverage summary, bean(s) closed.

Rules: never skip the gate; commit as Pim Snel, no self-promotion; keep changes
scoped to the one change.
