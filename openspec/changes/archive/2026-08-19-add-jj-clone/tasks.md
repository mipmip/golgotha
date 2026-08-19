## 1. Config: clone_vcs + vcs_rules + resolver

- [x] 1.1 Add `Config.CloneVCS \`yaml:"clone_vcs"\`` (global), `Provider.CloneVCS \`yaml:"clone_vcs"\``, and `Provider.VCSRules []VCSRule \`yaml:"vcs_rules"\`` with `VCSRule{ Match string \`yaml:"match"\`; VCS string \`yaml:"vcs"\` }`
- [x] 1.2 Add a pure `CloneVCSFor(p *Provider, ownerName string) string`: first matching `VCSRules` glob (`path.Match` on "owner/name") → provider `CloneVCS` → global `CloneVCS` → "git"
- [x] 1.3 Validate: every `vcs` is `git` or `jj`; every `match` compiles via `path.Match`; actionable errors
- [x] 1.4 Config unit tests: default git, per-repo rule wins, provider override, global default, invalid vcs, invalid glob
- [x] 1.5 Document `clone_vcs` and `vcs_rules` (match/vcs) in `config.example.yaml` (required by the config-example gate)

## 2. Syncer: colocated jj clone

- [x] 2.1 Branch the clone on `config.CloneVCSFor(...)` in `CloneRepo`/`CloneRepoProgress`/`syncRepo`: git → git clone (today); jj → `jj git clone --colocate`
- [x] 2.2 Guard `jj` on `PATH` before a jj clone; return an actionable error when missing
- [x] 2.3 Plain jj clone path (for `hup sync`): stderr piped, no pty, colocated
- [x] 2.4 Syncer unit tests: correct command selection per resolved vcs (fake runner); jj-missing error; colocated flag present

## 3. jj progress via pseudo-terminal

- [x] 3.1 Add a pty dependency (e.g. `creack/pty`); update `flake.nix` `vendorHash`
- [x] 3.2 PTY-progress jj clone: run `jj git clone --colocate --color never` on a pty, strip ANSI CSI sequences, split on `\r`, parse `(\d+)%` → fraction/phase events; spinner fallback on unparseable chunks; return the terminal result
- [x] 3.3 Unit tests: ANSI-strip + `NN%` parsing over representative jj output (captured in the spike); unparseable input yields no bogus percentage

## 4. Verification

- [x] 4.1 `gofmt -l .` empty; `go test ./...` passes; `nix flake check` passes (coverage gate)
- [x] 4.2 Keep the `skull2-wr68` beans checklist current as tasks complete
