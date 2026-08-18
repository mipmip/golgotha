---
# skull2-289n
title: alternative dir navigation using symlinks
status: draft
type: task
priority: normal
created_at: 2026-08-18T08:31:19Z
updated_at: 2026-08-18T08:44:15Z
parent: skull2-ok4c
---

My brain doesn't always think in forge organizations. Sometimes a project or personal topic has repo's spread over multiple organizations and maybe forges as well. I want a clean strategy to have virtual collections of repo's groups by different taxonomies.  e.g. Project X: org1/repo12, mipmip/repo3. And Topic Y, repox, repoz. Etc...

Some collections could be made automaticly: e.g. Lang Nix: .... Some should be made from inside the TUI. It should be able to maintain taxonomies from inside the app as well, crud functionaly.

I maintain my homefiles with Nix Home Manager. The storage of declared (non automatic) collections should be a seperate file from the main config. But it should be in the config dir. Or maybe there should be a layered storage in which the runtime created config overrides the default home manager config. This way I could create a upstream sync script in my up_home task uin my nixos conf.

- I think this virtual collections should be created using symlinks.
- automatic non declarible collections: forge-topics, language, archived, forked,...
- personal collections based stored in e.g. json
