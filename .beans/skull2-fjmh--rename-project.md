---
# skull2-fjmh
title: rename project
status: todo
type: task
priority: normal
created_at: 2026-08-17T18:43:01Z
updated_at: 2026-08-17T19:26:23Z
parent: skull2-qati
---

e.g. something with forge and multiple but lets chat more about this

also the cmd should be unique and if possible very short

## Decision

- **Project name:** `Golgotha`
- **CLI command:** `gol`

### Rationale

The codename `skull2` came from the image of *a skull holding all the
brains* — one local vault containing every repo from every forge. Golgotha
literally means **"the place of the skull"** (Aramaic *gulgōltā*), so it is
the most on-metaphor word available: your dead/archived/mirrored repos all
come to rest in one cranium. Cold and a little menacing without needing a
wink — the skull meaning does the work.

`gol` chosen as the command over the initialism `gg`: two-letter commands are
collision-prone (everyone's dotfiles alias `gg`, and it reads as "good game").
`gol` is short, essentially unclaimed, and derives directly from the brand.

### Scope of the rename (for implementation — not done yet)

- Go module path `github.com/mipmip/skull2` → new path
- Binary / cmd dir `cmd/skull2` → `cmd/gol` (command name `gol`)
- Nix `packages.default` name, `nix run` invocation, flake outputs
- Docs: `CLAUDE.md`, `BRIEFING.md`, `README`, `openspec/config.yaml` context
- Config env var prefix `SKULL2_<PROVIDER>_TOKEN` → new prefix
- Cache dir / any on-disk paths keyed on the old name
- Repo name on remote `git@github.com:mipmip/skull2.git` (open question:
  rename repo too, or keep repo slug and only rename the tool?)
