## 1. Version embedding

- [ ] 1.1 Create `VERSION` file (e.g. `0.1.0`, no `v` prefix)
- [ ] 1.2 Update `cmd/skull2/main.go` to embed VERSION via `go:embed` (keep the ldflags override)
- [ ] 1.3 Update `flake.nix` to read version via `builtins.readFile ./VERSION` (drop the hardcoded string)
- [ ] 1.4 Verify `go run ./cmd/skull2 version` and `nix run . -- version` both print the VERSION value

## 2. goreleaser

- [ ] 2.1 Create `.goreleaser.yaml`: builds for linux/amd64+arm64, darwin/amd64+arm64; binary `skull2`
- [ ] 2.2 Inject version from the git tag via ldflags; tar.gz archives + checksums
- [ ] 2.3 Add goreleaser + gum to the flake devShell
- [ ] 2.4 Verify `goreleaser check` and `goreleaser build --snapshot --clean`

## 3. GitHub Actions

- [ ] 3.1 Create `.github/workflows/release.yml` on `v*` tag push (checkout, Go setup, goreleaser action with GITHUB_TOKEN)

## 4. Release script

- [ ] 4.1 `scripts/release.sh` safety checks: clean tree, on main, CHANGELOG has Unreleased, tag absent, `nix flake check` passes
- [ ] 4.2 gum bump prompt (major/minor/patch); update VERSION; promote CHANGELOG Unreleased → version+date
- [ ] 4.3 vendorHash auto-update (fake hash → nix build → parse → write; skip with warning if nix absent)
- [ ] 4.4 jj-first commit + push (jj commit; bookmark main; jj git push) then `git tag vX.Y.Z @-` + `git push origin vX.Y.Z`
- [ ] 4.5 Make executable; keep it out of the coverage.* gitignore rule if needed

## 5. Docs

- [ ] 5.1 Create `CHANGELOG.md` (Keep a Changelog) with an `Unreleased` section
- [ ] 5.2 Create `RELEASING.md` (checklist, steps, verification, troubleshooting)

## 6. Verify

- [ ] 6.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
- [ ] 6.2 `goreleaser check` passes and a snapshot build produces all four targets
