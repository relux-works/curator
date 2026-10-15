# BUG-261015-npfg4c: fetch-rejects-repointed-upstream-tags

## Description
When an upstream repository re-points an existing release tag (for example after a history rewrite followed by a force-push of all tags), every fetch that goes through gitops.Fetch fails inside git: since git 2.20, fetch --all --tags --prune refuses to move an existing local tag without --force and prints ! [rejected] vX -> vX (would clobber existing tag). Reproduced 2026-10-15 with git 2.54 and the exact argv. Affected callers: curator update (cmd/curator/main.go cmdUpdate) and every env profile git source fetch (internal/envprofile/gitsource.go gitManager.fetch, used by profile update, install and overlays). The command fails with raw git output and never reaches moved-tag handling, so neither the default warning nor the strict-tag refusal that protocol/environments.md requires at profile update can happen. A search found no moved-tag detection in internal/envprofile at all. Not affected: the upgrade/install draft lane (gitops.FetchURLIsolated already uses --force with +refs/tags/*:refs/tags/*, then install.detectMovedTagsIn warns or fails under --strict-tags) and locked replay (fetch by object id, no tags).

## Scope
internal/gitops fetchWith/Fetch; curator update; internal/envprofile git source fetch and the profile update moved-tag diagnostic.

## Acceptance Criteria
1. A cached clone whose upstream re-pointed tags fetches successfully on both affected paths; no raw clobber error reaches the user. 2. profile update with a tag-pinned root whose tag moved: default prints a moved-tag warning naming tag, old and new commit; under strict-tag policy it fails with a typed diagnostic listing every moved tag, with no lock or store write. 3. Negative rows: an unchanged tag yields no diagnostic; a deleted upstream tag is handled explicitly; the fetch never moves a tag that did not move upstream. 4. Existing draft-lane and replay behaviour unchanged (regression rows kept).
