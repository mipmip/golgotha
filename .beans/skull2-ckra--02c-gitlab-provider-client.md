---
# skull2-ckra
title: 02c GitLab provider client
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:13:59Z
parent: skull2-6zs4
---

List repositories via the GitLab v4 API with group/subgroup nesting, using the `glab` token or `SKULL2_GITLAB_TOKEN`.

## Tasks
- [x] v4 list across groups/subgroups + owned projects, pagination
- [x] Auth via glab token, fallback env PAT; self-hosted `api_url`
- [x] Map to domain `Repo` (namespace -> owner); honor filters
- [x] Unit tests against mocked HTTP

## Summary of Changes

GitLab v4 client: groups with include_subgroups + membership projects, X-Next-Page pagination, PRIVATE-TOKEN auth, namespace->owner mapping, fork detection.
