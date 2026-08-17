## 1. Clone-path template engine

- [x] 1.1 Define the template data struct (BaseDir, Provider, Type, Short, Host, Owner, OwnerLower, Repo, RepoLower) in `internal/clonepath`
- [x] 1.2 Render with `text/template` and `missingkey=error`; resolve per-provider override vs global
- [x] 1.3 Expand `~`/base_dir, clean to absolute, and reject paths escaping base_dir
- [x] 1.4 Tests: default template TechNative example, field coverage, override, traversal rejection

## 2. Provider abstraction

- [x] 2.1 Define the `Repo` model in `internal/provider`
- [x] 2.2 Define the `Provider` interface (`ListRepos(ctx, owners) ([]Repo, error)`)
- [x] 2.3 Central archived/fork filtering helper honoring provider config
- [x] 2.4 Tests with a fake Provider (no network)

## 3. Auth resolution

- [x] 3.1 Auth resolver: configured CLI token → env PAT → actionable error
- [x] 3.2 Inject env lookup and CLI-token getter for hermetic tests
- [x] 3.3 Tests: cli token used, env fallback, missing-credential error

## 4. Registry

- [x] 4.1 Type→constructor registry with register + build-by-type
- [x] 4.2 Unknown type errors; tests using a registered fake type

## 5. Verify

- [x] 5.1 `go build ./...`, `go vet ./...`, `gofmt -l .` clean
- [x] 5.2 `nix flake check` passes
