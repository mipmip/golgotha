---
# skull2-thwi
title: 01d Provider & auth abstraction
status: todo
type: epic
priority: high
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-j0od
---

Provider interface and auth resolver shared by all concrete clients; no network in unit tests.

## Tasks
- [ ] `Provider` interface (e.g. ListRepos(ctx, owners) -> []Repo)
- [ ] `Repo` domain model (owner, name, urls, default_branch, archived, fork, updated_at)
- [ ] Auth resolver order: configured CLI token -> env PAT -> clear error
- [ ] Provider registry/factory keyed by type
- [ ] Unit tests with a fake provider (no network)
