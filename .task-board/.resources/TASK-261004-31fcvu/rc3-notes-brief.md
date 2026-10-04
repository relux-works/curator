# THE ONLY CURRENT INSTRUCTION — curator v0.15.0-rc.3 release notes (developer, docs)

Write the consolidated `## v0.15.0-rc.3` section of CHANGELOG.md for the tag the orchestrator will put on main right after this lands. Source of truth: the merged history `git log v0.15.0-rc.2..origin/main` (read it in full; the history was re-identified on 2026-10-03, so compare by subject/content, not by old hashes) plus the readiness plan `.research/261002_rc14_rc3_readiness.md` § "Curator rc.3: payload, pin and release gates".

## Do
1. Turn `## Unreleased` into `## v0.15.0-rc.3 — 2026-10-04` and leave a fresh empty `## Unreleased` above it. Group as Added / Changed / Fixed / Security / Known issues, following the existing style of earlier release sections.
2. Reconcile against history: include work that landed after rc.2 but has NO Unreleased line yet, at least:
   - the muse environment adapter and launch-env-fragment v3 (curator#100, curator-spec#121);
   - launch-env-fragment v2 with 0018 permission fragments, credential migrate and `env unmanage` restore;
   - the global-upgrade cache sweep no longer deleting binaries of live processes (TASK-261002-ot3ea1);
   - the Windows executable hard-link origin validation, the marker v3/v4 cross-field validation, the Unix askpass EPIPE fix (already listed: keep, de-duplicate);
   - the conformance pin advanced to curator-spec 1.0.0-rc.14 (final tag commit 43bf0a25, manifest 6f832d81); hashes still WRITTEN as curator-content-v1, v2 read-compatible; the one rc.14 v2-write case is an owned known gap (TASK-261003-1uzji7, rc.4);
   - Codex seed revision A and machine security-posture revision A shipping for the first time in a tagged warning release.
   Remove the stale "keep rc.13 pinned" wording. Do not claim anything not in history.
3. Known issues (state plainly, no "fixed" claims):
   - Windows broker real-Git flake TASK-260930-fp8vx7: two historical failures, not reproduced in ~21,000 hosted passes, no root cause; shipped as a documented risk.
   - v2 hash writing is deferred to rc.4 (atomic v1→v2 migration first).
   - The 2026-10 inline security audit (issue #106, docs/security-audit-2026-10-inline.md) found N1–N4; they are NOT fixed in rc.3 (tracked in STORY-261004-3oognx).
   - B3 cache-prune PRs are not included.
4. Only CHANGELOG.md changes. Never edit LOGBOOK.md. No version constants exist (GoReleaser ldflags set the version).

Evidence in the outcome: the commit-subject list you reconciled against and which entries map to which commits. Then `task-board handoff <TASK> --role developer` and END YOUR TURN.

## Update (orchestrator): start from the saved draft
Two earlier runs stalled on host command execution after drafting. Their draft is attached as `rc3-notes-draft-run1.md` (a full `## Unreleased` + `## v0.15.0-rc.3` section). Start from it: verify every line against the history, fix what is wrong, then apply it to CHANGELOG.md. Keep commands short and few; if a command hangs for more than 5 minutes, wait (host rules: syspolicyd) rather than retrying in a loop.
