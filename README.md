# huphop

Multi-provider git portfolio manager — a uniform on-disk layout for all your
git repos across GitHub, Codeberg/Forgejo and GitLab, a TUI to browse and
clone them, and a cron-friendly CLI to keep backup clones in sync.

> Status: PoC / alpha base. See [BRIEFING.md](./BRIEFING.md) for the full
> design and [CLAUDE.md](./CLAUDE.md) for the build guide. The roadmap lives in
> `beans` (`beans list`).

## Quick start

```bash
nix develop            # dev shell
nix build              # build
nix run . -- version   # run
nix flake check        # build + tests
```

## What it does

- **Uniform layout** via a configurable clone-path template, e.g. GitHub
  `TechNative-B-V/foo` → `~/gh.technative-b-v/foo`.
- **TUI**: browse provider → owner → repos with fuzzy search; single/bulk
  clone; open in browser.
- **CLI sync**: clone missing repos and fast-forward-pull existing ones;
  dirty-tree safe, cron-friendly.

Configuration lives in `~/.config/huphop/config.yaml`; cached repo metadata in
`~/.cache/huphop/`.
