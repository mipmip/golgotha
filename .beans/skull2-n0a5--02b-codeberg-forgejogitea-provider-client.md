---
# skull2-n0a5
title: 02b Codeberg (Forgejo/Gitea) provider client
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:13:59Z
parent: skull2-6zs4
---

List repositories via the Forgejo/Gitea REST API using env PAT `SKULL2_CODEBERG_TOKEN`.

## Tasks
- [x] REST list for user + orgs, full pagination
- [x] Env-PAT auth; configurable `api_url`
- [x] Map to domain `Repo`; honor filters
- [x] Unit tests against mocked HTTP

## Summary of Changes

Codeberg (Forgejo/Gitea) client: /api/v1 with page/limit pagination, token auth, configurable api_url.
