## 1. Remote and module

- [ ] 1.1 Rename the GitHub repo `mipmip/skull2` → `mipmip/golgotha` and update
      the local `jj`/git remote URL to `git@github.com:mipmip/golgotha.git`
- [ ] 1.2 Change the module path in `go.mod` to `github.com/mipmip/golgotha`
- [ ] 1.3 Rename the entrypoint directory `cmd/skull2` → `cmd/gol`

## 2. Code rename

- [ ] 2.1 Rewrite every `github.com/mipmip/skull2/...` import to
      `github.com/mipmip/golgotha/...` across all `.go` files (incl. tests
      and `e2e/e2e_test.go`)
- [ ] 2.2 Rename XDG config path segment `skull2` → `golgotha` in
      `internal/config/config.go` (`~/.config/golgotha/config.yaml`) and update
      its doc comments
- [ ] 2.3 Rename XDG cache path segment `skull2` → `golgotha` in
      `internal/cache/cache.go` (`~/.cache/golgotha`) and update doc comments
- [ ] 2.4 Replace the auth env-var prefix `SKULL2_` → `GOLGOTHA_` in
      `internal/provider` and every `*_test.go` fixture that sets or references
      `SKULL2_<PROVIDER>_TOKEN`
- [ ] 2.5 Update the CLI binary/usage name and any `skull2`-branded strings in
      `cmd/gol/main.go` (and `main_test.go`) so help/version print `gol`

## 3. Packaging and tooling

- [ ] 3.1 Update `flake.nix`: `subPackages = [ "cmd/gol" ]` and package/binary
      name `gol`; confirm `nix build` and `nix run . -- version` still work
- [ ] 3.2 Update core-package import paths in `scripts/coverage.sh` to the new
      module path
- [ ] 3.3 Update `.beans.yml` and `config.example.yaml` name/env references
- [ ] 3.4 If `nix build` reports a `vendorHash` mismatch, update it from the
      printed expected hash (no deps changed, so likely unchanged)

## 4. Docs

- [ ] 4.1 Rebrand `README.md` title and body to Golgotha / `gol`
- [ ] 4.2 Rebrand `BRIEFING.md` (name, env vars, paths, remote URL)
- [ ] 4.3 Rebrand `CLAUDE.md` (module path, `cmd/gol`, env prefix, remote,
      commands)
- [ ] 4.4 Rebrand `openspec/config.yaml` `context:` block

## 5. Verify

- [ ] 5.1 `gofmt -l .` returns empty
- [ ] 5.2 `nix flake check` passes (build + tests + coverage gate)
- [ ] 5.3 Residue sweep: `grep -rIn 'skull2\|SKULL2\|Skull2'` across the tree
      (excluding `.git` and `openspec/changes/archive`) returns no hits in
      code, packaging, tooling, or active docs
