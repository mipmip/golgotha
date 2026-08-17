---
# skull2-gd2b
title: 02a GitHub provider client
status: todo
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-6zs4
---

List repositories via the GitHub REST API with pagination and auth from the `gh` CLI token or `SKULL2_GITHUB_TOKEN`.

## Tasks
- [ ] REST list for user + configured orgs, full pagination
- [ ] Auth via gh CLI token, fallback env PAT; GHE `api_url` override
- [ ] Map to domain `Repo`; honor include_archived/include_forks
- [ ] Unit tests against recorded/mocked HTTP
