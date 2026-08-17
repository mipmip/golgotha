---
# skull2-ckra
title: 02c GitLab provider client
status: todo
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T08:39:28Z
parent: skull2-6zs4
---

List repositories via the GitLab v4 API with group/subgroup nesting, using the `glab` token or `SKULL2_GITLAB_TOKEN`.

## Tasks
- [ ] v4 list across groups/subgroups + owned projects, pagination
- [ ] Auth via glab token, fallback env PAT; self-hosted `api_url`
- [ ] Map to domain `Repo` (namespace -> owner); honor filters
- [ ] Unit tests against mocked HTTP
