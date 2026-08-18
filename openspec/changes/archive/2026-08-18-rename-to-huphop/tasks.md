## 1. Remote & module

- [x] 1.1 Repoint local `jj`/git remote to `git@github.com:mipmip/huphop.git` (GitHub repo already renamed)
- [x] 1.2 `go.mod` module → `github.com/mipmip/huphop`; rewrite every import across all `*.go` (incl. tests + `e2e`)
- [x] 1.3 `git mv cmd/gol cmd/hup`; root `version.go` package `golgotha` → `huphop`; update its importer

## 2. Code rename

- [x] 2.1 CLI/brand strings in `cmd/hup/main.go` (+ tests): help/usage/version print `hup`
- [x] 2.2 Config dir `~/.config/golgotha` → `~/.config/huphop` (`internal/config` + comments + fixtures)
- [x] 2.3 Cache dir `~/.cache/golgotha` → `~/.cache/huphop` (`internal/cache` + comments + fixtures)
- [x] 2.4 Env prefix `GOLGOTHA_` → `HUPHOP_` in `internal/provider` + all `*_test.go`/e2e

## 3. Packaging & tooling

- [x] 3.1 `flake.nix`: `pname`/`mainProgram` `hup`, `subPackages = ["cmd/hup"]`; keep vendorHash, VERSION readFile, doCheck, checks
- [x] 3.2 `.goreleaser.yaml` (binary `hup`, main `./cmd/hup`, project name), `.github/workflows/release.yml`
- [x] 3.3 `scripts/coverage.sh` core import paths → `github.com/mipmip/huphop/...`; `scripts/release.sh` brand/URL strings
- [x] 3.4 If `nix build` reports a vendorHash mismatch, update from the printed hash (no deps changed, likely unchanged)

## 4. Docs & specs

- [x] 4.1 Rebrand `README.md`, `BRIEFING.md`, `CLAUDE.md`, `RELEASING.md`, `CHANGELOG.md`, `config.example.yaml`, `openspec/config.yaml` context
- [x] 4.2 Sweep `golgotha`/`GOLGOTHA_` in the non-delta main specs only: `openspec/specs/{repo-cache,sync,version-embedding}/spec.md` (do NOT touch configuration/provider-clients main specs — the deltas handle those; do NOT touch `openspec/changes/archive/`)
- [x] 4.3 Leave `.beans.yml` `prefix` as `skull2-`; do not rename existing `.beans/` files

## 5. Verify

- [x] 5.1 No stray refs: `grep -rE "golgotha|GOLGOTHA_|cmd/gol|\bgol\b" --include=*.go .` empty; likewise flake.nix/.goreleaser.yaml/scripts/docs
- [x] 5.2 `gofmt -l .`, `go vet`, `go build`, `go test ./...` pass; `go run ./cmd/hup version` prints `hup <version>`
- [x] 5.3 `nix build` + `./result/bin/hup version` prints `hup <version>`; `bash scripts/coverage.sh` PASS; `nix flake check` green
