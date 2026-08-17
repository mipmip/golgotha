## 1. GitHub client

- [ ] 1.1 `internal/provider/github.go`: list user + org repos via REST with pagination
- [ ] 1.2 Auth (gh token / SKULL2_GITHUB_TOKEN), `api_url` override; map to Repo
- [ ] 1.3 Register `github` in the registry
- [ ] 1.4 Tests against httptest with paginated fixtures

## 2. Codeberg client

- [ ] 2.1 `internal/provider/codeberg.go`: Forgejo/Gitea REST list with pagination
- [ ] 2.2 Env-PAT auth, configurable `api_url`; map to Repo
- [ ] 2.3 Register `codeberg`; tests against httptest

## 3. GitLab client

- [ ] 3.1 `internal/provider/gitlab.go`: v4 group/subgroup + owned projects, pagination
- [ ] 3.2 Auth (glab token / SKULL2_GITLAB_TOKEN), self-hosted `api_url`; map namespace→owner
- [ ] 3.3 Register `gitlab`; tests against httptest

## 4. Cache & refresh

- [ ] 4.1 `internal/cache`: schema (fetched_at + repos), atomic write, read, path (`$XDG_CACHE_HOME` or ~/.cache/skull2)
- [ ] 4.2 `skull2 refresh [--provider NAME]` re-fetches and writes cache; unknown provider exits non-zero
- [ ] 4.3 Tests: round-trip, atomic write, missing cache, refresh wiring

## 5. Verify

- [ ] 5.1 `gofmt -l .`, `go vet ./...`, `go build ./...` clean
- [ ] 5.2 `go test ./...` passes; core packages >=80% coverage
- [ ] 5.3 `nix flake check` passes
