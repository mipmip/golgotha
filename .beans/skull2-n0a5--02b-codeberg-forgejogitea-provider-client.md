---
# skull2-n0a5
title: 02b Codeberg (Forgejo/Gitea) provider client
status: todo
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-6zs4
---

List repositories via the Forgejo/Gitea REST API using env PAT `SKULL2_CODEBERG_TOKEN`.

## Tasks
- [ ] REST list for user + orgs, full pagination
- [ ] Env-PAT auth; configurable `api_url`
- [ ] Map to domain `Repo`; honor filters
- [ ] Unit tests against mocked HTTP
