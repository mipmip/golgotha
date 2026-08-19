## 1. Config: mandatory username, delete sentinel

- [x] 1.1 Add `Username string \`yaml:"username"\`` to `config.Provider` with a doc comment
- [x] 1.2 Extend `Config.Validate()` to reject an empty `username` per provider with an actionable error (alongside `name`/`type`/`short`)
- [x] 1.3 Remove `config.SelfOwner` and `selfExcludeToken` and all references to them
- [x] 1.4 Rewrite `ResolveOwners`: empty `owners:` → `[username]`; `all_owners` unions `username` + discovered orgs + explicit owners; `exclude_owners` matches `username` case-insensitively; drop the `""` prepend and the `"self"` token
- [x] 1.5 Update config unit tests and the documented example config for the new required field; adjust `ResolveOwners` tests to expect real logins

## 2. Provider clients: route self fetch by username

- [x] 2.1 GitHub `FetchOwner`: choose `/user/repos?affiliation=owner` when `owner == cfg.Username`, else `/orgs/<owner>/repos`; remove the `SelfOwner` branch
- [x] 2.2 GitLab `FetchOwner`: route the self (authenticated-user) listing when `owner == cfg.Username`, else group/subgroup; remove the `SelfOwner` branch
- [x] 2.3 Codeberg `FetchOwner`: same routing keyed on `cfg.Username`; remove the `SelfOwner` branch
- [x] 2.4 Update provider unit tests (mock HTTP) to cover self-owner routing by configured username and the org path; confirm fetched self repos carry the real login

## 3. Cache & CLI: drop sentinel handling

- [x] 3.1 Remove the `SelfOwner`/`""` support note and any `""` fixtures from `internal/cache` (and its tests)
- [x] 3.2 `cmd/hup`: remove `displayOwner`/`SelfOwner` handling; default the owner set to `[username]`; update `progress.go` and tests

## 4. TUI: pin + tint self owner, fix scoping

- [x] 4.1 `ownersFor`: pin the owner equal to `provider.Username` first; keep the rest sorted
- [x] 4.2 Remove `ownerLabel`'s `SelfOwner` → `"(your account)"` mapping; show the real login
- [x] 4.3 Apply a distinct lipgloss style to the self-owner row in the owner-level view
- [x] 4.4 Remove the `selOwner != ""` guard in `visibleRepos()` so the self owner scopes to its own repos
- [x] 4.5 Update TUI unit tests: self owner pinned first, distinguished, labeled by login, and scoping to only its repos

## 5. Docs & verification

- [x] 5.1 Update `BRIEFING.md` and `openspec/config.yaml` `context:` to document the mandatory `username` field
- [x] 5.2 Update `config.example.yaml`: add `username` to the provider, drop the "leave owners unset for your own repos" guidance and the `self` exclude token (align it with the new model so `hup config check` passes)
- [x] 5.3 `gofmt -l .` is empty and `nix flake check` passes (build + tests + coverage gate)
- [x] 5.4 Keep the `skull2-qp5y` beans checklist current as tasks complete
