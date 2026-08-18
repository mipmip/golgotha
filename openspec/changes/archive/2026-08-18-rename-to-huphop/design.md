## Context

Second wholesale rename (skull2 → golgotha → huphop), using the proven
`rename-to-golgotha` playbook. This time the release-process files
(`.goreleaser.yaml`, `release.sh`, workflow, `RELEASING.md`, `CHANGELOG.md`,
root `version.go`) already exist and are in scope from the start. The GitHub
repo is already `mipmip/huphop`; only the local remote needs repointing.

## Goals / Non-Goals

**Goals:** complete, gate-green rebrand golgotha→huphop / gol→hup with no stray
references in code/tooling/active docs.

**Non-Goals:** behavior changes; touching archived changes; renaming existing
`.beans/` files or the bean ID prefix; a data-migration for existing
`~/.config/golgotha` / `~/.cache/golgotha` (documented manual step).

## Decisions

- **Command `hup`** (not `hh`): three letters, unique, no `hstr` collision,
  greppable.
- **Rename map:** module `github.com/mipmip/golgotha`→`…/huphop`; dir
  `cmd/gol`→`cmd/hup`; binary/CLI `gol`→`hup`; env `GOLGOTHA_`→`HUPHOP_`; config
  `~/.config/golgotha`→`~/.config/huphop`; cache `~/.cache/golgotha`→
  `~/.cache/huphop`; root package `golgotha`→`huphop`.
- **Remote:** repoint `origin` to `git@github.com:mipmip/huphop.git` (repo
  already renamed) as the first step so the ship push lands.
- **vendorHash:** keep the existing value; module rename adds no dependencies.
  If nix reports a mismatch, recompute from the printed hash.
- **Spec deltas:** this change carries `configuration` + `provider-clients`
  deltas (env/path). Golgotha references in the other main specs (`repo-cache`,
  `sync`, `version-embedding`) are swept directly as doc-only edits (not delta
  files) to avoid archive conflicts.
- **`.beans.yml` prefix** stays `skull2-` (renaming it would orphan existing
  bean IDs; internal-only, not product identity).

## Risks / Trade-offs

- [Third rename / churn] → accepted; name declared final. Playbook is proven and
  gated by `nix flake check`.
- [Stale local config/cache under golgotha] → not auto-migrated; note it in docs.
- [Two-letter vs three-letter command] → `hup` chosen to avoid collisions.
