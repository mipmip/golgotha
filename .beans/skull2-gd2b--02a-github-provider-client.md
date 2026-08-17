---
# skull2-gd2b
title: 02a GitHub provider client
status: completed
type: epic
priority: normal
created_at: 2026-08-17T08:39:28Z
updated_at: 2026-08-17T09:13:59Z
parent: skull2-6zs4
---

List repositories via the GitHub REST API with pagination and auth from the `gh` CLI token or `SKULL2_GITHUB_TOKEN`.

## Tasks
- [x] REST list for user + configured orgs, full pagination
- [x] Auth via gh CLI token, fallback env PAT; GHE `api_url` override
- [x] Map to domain `Repo`; honor include_archived/include_forks
- [x] Unit tests against recorded/mocked HTTP

## Summary of Changes

GitHub REST client: Link-header pagination, Bearer auth, GHE api_url override, user + org listing, mapped to Repo with archived/fork filtering.
