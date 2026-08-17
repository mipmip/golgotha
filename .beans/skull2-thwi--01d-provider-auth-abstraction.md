---
# skull2-thwi
title: 01d Provider & auth abstraction
status: completed
type: epic
priority: high
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:06:02Z
parent: skull2-j0od
---

Provider interface and auth resolver shared by all concrete clients; no network in unit tests.

## Tasks
- [x] `Provider` interface (e.g. ListRepos(ctx, owners) -> []Repo)
- [x] `Repo` domain model (owner, name, urls, default_branch, archived, fork, updated_at)
- [x] Auth resolver order: configured CLI token -> env PAT -> clear error
- [x] Provider registry/factory keyed by type
- [x] Unit tests with a fake provider (no network)

## Summary of Changes

Implemented `internal/provider`: Repo model, Provider interface, archived/fork filtering, auth resolver (CLI token -> env PAT -> error, injectable), and type->constructor registry. Coverage 86%.
