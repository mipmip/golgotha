# HupHop — PoC Briefing

> This document is the single source of truth for autonomously building the
> HupHop PoC (alpha base). It is written to be consumed by `mip:1shotpoc`.

## 1. Purpose

HupHop controls an ever-growing portfolio of git projects spread across
multiple git providers (GitHub, Codeberg, GitLab today; Bitbucket later). It
gives a **uniform on-disk directory structure**, a **browsable TUI** for
discovering and cloning repos, and a **headless CLI** for cron-driven backup
syncing. Configuration plus cacheable remote data form a single source of
truth.

## 2. Goals

- Uniform directory structure aligned with git-provider structure.
- Enable derived tmux configuration for quick navigation (structure only in
  PoC; generator is a later stretch).
- Let the existing `dirtygit` util discover dirty repos (structure enables it).
- Single source of truth: YAML config + cacheable remote repo data.
- Interactive browsing (TUI).
- Cron-friendly auto-syncing for backups (CLI).

## 3. Non-goals

- Code search.
- tmux config generator (structure is enough for PoC; generator is later).
- Providers beyond GitHub, Codeberg and GitLab (design leaves room; do not
  build Bitbucket etc. yet).

## 4. Decisions (locked)

| Topic       | Decision                                                          |
|-------------|-------------------------------------------------------------------|
| Language    | Go, Bubble Tea (+ bubbles, lipgloss) for the TUI                  |
| Providers   | GitHub, Codeberg (Forgejo/Gitea), GitLab                         |
| Auth        | Reuse `gh` / `glab` CLIs; PAT via env var fallback for CLI/cron   |
| Sync mode   | Clone missing + fast-forward pull existing (working trees)        |
| Cache       | Plain JSON file per provider under `~/.cache/huphop/`             |
| Clone path  | Configurable Go text/template (`clone_pattern_tpl`) with fields   |
| Coverage    | Enforced min **70%** overall; **≥80%** on core logic packages    |
| PoC scope   | TUI browse (fuzzy + hierarchic), clone single/bulk, open browser, CLI sync |
| VCS         | `jj`; remote `git@github.com:mipmip/huphop.git`; commit as Pim Snel, no self-promo |
| Tickets     | `beans` milestones/epics, milestone titles prefixed `01`, `02`... |
| Specs       | OpenSpec proposals + tasks, fully set up before build            |
| Packaging   | Nix flakes, plain nix (no flake-utils), multi-arch               |
| Testing     | Thorough unit tests + e2e cases proving the PoC works            |

## 5. Directory / clone path template

The clone target path is a **configurable Go `text/template`**, not a fixed
scheme. This lets extra characters, separators, and the base dir be arranged
freely. Confirmed baseline from Pim's own filesystem (`~/gh.mipmip/...`):

Default template:

```
{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}
```

Available data fields:

| Field         | Meaning                                   | Example              |
|---------------|-------------------------------------------|----------------------|
| `.BaseDir`    | Configured base dir (expanded `~`)        | `/home/pim`          |
| `.Provider`   | Provider `name`                           | `github-personal`    |
| `.Type`       | Provider `type`                           | `github`             |
| `.Short`      | Provider short code                       | `gh`                 |
| `.Host`       | Provider host                             | `github.com`         |
| `.Owner`      | Owner/org, upstream casing                | `TechNative-B-V`     |
| `.OwnerLower` | Owner/org lowercased                      | `technative-b-v`     |
| `.Repo`       | Repo name, upstream casing                | `Foo`                |
| `.RepoLower`  | Repo name lowercased                      | `foo`                |

Example: GitHub org `TechNative-B-V` repo `foo` → `~/gh.technative-b-v/foo`.
Provider short codes: `gh` (GitHub), `cb` (Codeberg), `gl` (GitLab).

## 6. Configuration

Location: `~/.config/huphop/config.yaml`. Proposed schema (Claude may refine
during OpenSpec, keep it minimal and documented):

```yaml
base_dir: ~/                       # root for all trees; default ~

# Global default clone path template (Go text/template). Fields in §5.
# Per-provider `clone_pattern_tpl` overrides this.
clone_pattern_tpl: "{{.BaseDir}}/{{.Short}}.{{.OwnerLower}}/{{.Repo}}"

providers:
  - name: github-personal          # unique key
    type: github                   # github | codeberg | gitlab
    short: gh                       # path prefix
    api_url: https://api.github.com # override for GHE
    web_url: https://github.com     # for "open in browser"
    clone_protocol: ssh             # ssh | https
    # clone_pattern_tpl: "..."      # optional per-provider override
    auth:
      cli: gh                       # reuse this CLI's token when present
      env: HUPHOP_GITHUB_TOKEN      # PAT fallback (headless/cron)
    owners:                         # optional allow-list; empty = all accessible
      - mipmip
      - TechNative-B-V
    include_archived: false         # skip archived repos by default
    include_forks: true

  - name: codeberg
    type: codeberg
    short: cb
    api_url: https://codeberg.org   # Forgejo/Gitea REST API base
    web_url: https://codeberg.org
    clone_protocol: ssh
    auth:
      env: HUPHOP_CODEBERG_TOKEN    # env-PAT first (no first-party CLI)

  - name: gitlab
    type: gitlab
    short: gl
    api_url: https://gitlab.com/api/v4  # override for self-hosted
    web_url: https://gitlab.com
    clone_protocol: ssh
    auth:
      cli: glab                     # reuse glab token when present
      env: HUPHOP_GITLAB_TOKEN      # PAT fallback (headless/cron)
```

- Auth resolution order per provider: configured `cli` token → `env` PAT →
  error with a clear message.
- Codeberg uses the Forgejo/Gitea REST API; GitLab uses the v4 REST API
  (respect group/subgroup nesting for owners).

## 7. Cache

- One JSON file per provider: `~/.cache/huphop/<provider-name>.json`.
- Contents: `fetched_at` timestamp + list of repos with: owner, name,
  description, ssh_url, https_url, web_url, default_branch, archived, fork,
  updated_at.
- TUI reads cache; a refresh action re-fetches from the API. `hup sync`
  refreshes then acts.

## 8. Commands (CLI surface)

- `hup` / `hup tui` — launch the TUI (default).
- `hup sync [--provider NAME] [--no-refresh]` — headless: refresh cache,
  then clone-missing + ff-pull-existing across configured owners. Cron target.
- `hup refresh [--provider NAME]` — refresh cache only.
- `hup config path|check` — show config path / validate config + auth.

Exit codes and non-interactive output must be cron-friendly (structured,
line-oriented logging; non-zero on failure).

## 9. TUI behaviour

- Hierarchic navigation: **provider → owner/org → repos**.
- Global **fuzzy filter** (`/`) across the flattened repo list.
- **Single select** → clone to configured target pattern.
- **Multi select** (space) → bulk clone.
- **Open in browser** (`o`) → provider `web_url` for the repo.
- **Refresh** (`r`) → re-fetch current provider's cache.
- Clear status feedback: already-cloned vs not; clone/pull progress; errors.
- Keybindings shown in a help/footer bar; `q` quits.

## 10. Sync semantics (backup)

For each configured owner's repos:
- If target dir absent → clone using `clone_protocol` at the templated path.
- If present and a git repo → `git fetch` + fast-forward-only pull on the
  default branch; never force, never touch dirty trees (skip + warn).
- Report a per-provider summary (cloned / updated / skipped / failed).

## 11. Milestones & epics (beans + OpenSpec)

Milestone titles are two-digit prefixed. Suggested breakdown (Claude may adjust
granularity but keep this order and gating):

- **01 Foundations** — repo scaffold, nix flake (plain, multi-arch), Go module,
  config loader + validation, clone-path template engine, provider/auth
  abstraction.
- **02 Provider clients & cache** — GitHub, Codeberg and GitLab listing
  (pagination, auth resolution, GitLab subgroups), JSON cache read/write,
  `refresh`.
- **03 CLI sync** — clone-missing + ff-pull-existing, dirty-tree safety, cron
  friendly output, `sync`.
- **04 TUI** — hierarchic browse + fuzzy filter, single/bulk clone, open in
  browser, refresh.
- **05 Testing, e2e & coverage** — unit coverage for config/template/cache/sync
  logic; e2e cases (mocked provider APIs + temp dirs) proving browse→clone and
  sync flows; enforce the coverage gate (§12).

Each OpenSpec change is committed with `jj` after archival. Beans administers
milestones/epics throughout.

## 12. Coverage gate

- Overall project coverage **≥ 70%** (reasonable for a PoC with a TUI layer).
- Core-logic packages (config, clone-path template, cache, sync, provider
  clients) **≥ 80%**.
- TUI rendering code is exempt from the strict threshold but must have at least
  smoke/update-function tests.
- Coverage is measured in the test suite and enforced in the `nix flake check`
  / CI-style invocation; a build fails below threshold.

## 13. Definition of done (PoC)

- `nix build` / `nix run` work on supported architectures via the flake.
- `hup sync` clones and updates a configured tree end-to-end (proven by e2e).
- TUI can browse all three providers, fuzzy-find, single/bulk clone to the
  correct templated target paths, and open a repo in the browser.
- Config + JSON cache act as the single source of truth.
- Unit + e2e suites pass and coverage meets §12 under `nix flake check`.

## 14. Integration notes (non-PoC, keep compatible)

- Directory layout must remain compatible with `dirtygit` discovery.
- Layout should be trivially convertible into tmux session definitions later.
