## 1. Provider detail + README

- [x] 1.1 Add `RepoDetails(ctx, owner, name)` (stars/topics/language) and `Readme(ctx, owner, name)` to the Provider interface (update fakes)
- [x] 1.2 Implement for GitHub (`stargazers_count`/`topics`/`/readme`)
- [x] 1.3 Implement for Codeberg/Gitea (stars/topics/raw README)
- [x] 1.4 Implement for GitLab (`star_count`/topics/files README)
- [x] 1.5 Client tests against mocked HTTP (details + README + not-found)

## 2. Detail cache

- [x] 2.1 Per-repo detail cache `details/<provider>/<owner>__<repo>.json` { fetched_at, stars, topics, language, readme_md }; atomic write; separate from list cache
- [x] 2.2 Load/save/refresh helpers; store raw markdown
- [x] 2.3 Cache tests (round-trip, missing, refresh)

## 3. Rendering

- [x] 3.1 Add `charmbracelet/glamour`; render raw markdown at view width (flake.nix vendorHash left for the orchestrator per instructions)
- [x] 3.2 README in a `bubbles/viewport` (scrollable)

## 4. TUI detail view

- [x] 4.1 New detail navigation level; Enter opens it, `c` is the sole clone key; Esc returns to the repo list at prior position
- [x] 4.2 Header (description/stars/topics/language/updated/visibility) + scrollable README; loading indicator on first open
- [x] 4.3 Lazy fetch via tea.Cmd → detailLoadedMsg; `r` re-fetches; graceful offline (metadata + "README unavailable", no error screen)
- [x] 4.4 Update-driven tests (enter opens; lazy fetch populates+caches; cached reuse; offline fallback; esc back)

## 5. Verify

- [x] 5.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
