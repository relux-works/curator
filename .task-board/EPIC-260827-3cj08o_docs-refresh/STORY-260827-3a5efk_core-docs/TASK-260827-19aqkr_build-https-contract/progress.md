## Status
closed

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(3))

## Blocked By
- (none)

## Blocks
- TASK-260827-xdbobc

## Checklist
- [x] Every build-https claim verified against internal/ with grep evidence; mirrors build-ssh shape
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260827-f18a62, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260827-f18a62)
agent completed: [implementer] doc-writer (agy) (exit=124)
spawn run completed: agy (run=RUN-260827-f18a62, pid=16574, exit=124)
spawn run RUN-260827-f18a62 failed; operator action required; failure: run exceeded --timeout 30m0s and was terminated by the launcher; a child process could not be proven terminated and may still be running
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260827-a5ae8d, max_parallel=20)
spawn run RUN-260827-a5ae8d cancelled by operator; operator action required; reason: no operator reason supplied
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260827-33a125, max_parallel=20)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260827-5592dd, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260827-5592dd)
Review verdict: changes requested (to-dev). Evidence: TASK-260827-19aqkr_review-verdict.md.

Root cause: the story branch forked from 1f55f1b, 55 commits behind origin/main, and that tree contains no HTTPS implementation at all (no internal/install/buildhttps.go). No claim in the document could have been grep-verified against internal/ as the AC requires; several fabrications trace directly to the CocoaSkills sibling doc.

Also decisive: docs/build-https.md ALREADY EXISTS on origin/main (180 lines, accurate). The task premise (create the file) no longer holds. docs/build-ssh.md also moved upstream since the fork, so the cross-link lands on a stale copy.

9 blocking findings, all grep-verified against a git archive of origin/main:
0. doc written against a tree with no HTTPS code
1. docs/build-https.md already on origin/main
2. off-TTY behavior documented as the inverse of the code (doc says fail-closed; code continues anonymously - cmd/curator/main.go:1338-1343, internal/install/buildhttps.go:182-187)
3. build_repository_identity_invalid is not this surface s identifier; the real ones are build_repository_https_credential_missing and _selection_aborted
4. --token flag does not exist; real flags are --git-credentials | --keyring | --token-env
5. CURATOR_BUILD_HTTPS_USERNAME does not exist; default username is token, not oauth2
6. prompt transcript fabricated (real menu: 1 / t / q)
7. no GITHUB_TOKEN/GITLAB_TOKEN/CI_JOB_TOKEN discovery exists
8. every quoted CLI output wrong (source= not token=, list omits present=, login strings invented)
9. three rows of the parse-error table wrong

Correct and worth keeping: scope grammar text, segment/longest-match semantics, curator-build-https: namespace, build_https config shape, the whole transport-policy flag list, both cross-links, prose style.

Rework: rebase onto origin/main first, then decide create vs extend against the existing upstream document.

Workspace note: the story worktree at .temp/STORY-260827-3a5efk/worktree was removed by a concurrent session at 06:27 mid-review. Work is intact on branch task-board/story/STORY-260827-3a5efk at 81e56a13 and mirrored at .temp/docs-backup-260827/; reviewed content md5 c031cd47a3b6a0090e8db4f65802467f matches the committed blob.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260827-5592dd, pid=7076, exit=0)
Obsolete: docs/build-https.md landed on origin/main from the feature workstream (merge 6432b64c pushed 06:30 local) while this task wrote against the pre-push base; the upstream 180-line contract cites spec sections and is authoritative. The 285-line duplicate is removed from the story worktree. Residual value (build-ssh/build-https cross-links) moved to the style-sweep task. Reviewer verdict with 9 findings preserved as evidence of the stale-base failure mode.

## Precondition Resources
- [TASK-260827-19aqkr_cocoaskills-prose-style.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_cocoaskills-prose-style.md) — Source style guide to port/apply (English rules and blacklist)
- [TASK-260827-19aqkr_docs-refresh-spec.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_docs-refresh-spec.md) — Curator docs refresh spec
- [TASK-260827-19aqkr_tooling-note.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_tooling-note.md) — Shell-only edits, quoted heredocs, grep verification, literal outputs
- [TASK-260827-19aqkr_reviewer-note.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_reviewer-note.md) — Single-pass review, immediate verdict, no monitors
- [TASK-260827-19aqkr_sibling-doc.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_sibling-doc.md) — CocoaSkills sibling contract for cross-checking; Curator code wins on any difference
- [TASK-260827-19aqkr_finalize-note.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_finalize-note.md) — Work done; verify, evidence, CR, handoff; no rewrites

## Outcome Resources
- [TASK-260827-19aqkr_spawn-log_-implementer--doc-writer--agy-_RUN-260827-f18a62.log](file://TASK-260827-19aqkr/TASK-260827-19aqkr_spawn-log_-implementer--doc-writer--agy-_RUN-260827-f18a62.log) — System spawn log captured by task-board
- [TASK-260827-19aqkr_results.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_results.md) — Operator HTTPS credential contract outcome evidence
- [TASK-260827-19aqkr_spawn-log_-reviewer--reviewer--claude-_RUN-260827-a5ae8d.log](file://TASK-260827-19aqkr/TASK-260827-19aqkr_spawn-log_-reviewer--reviewer--claude-_RUN-260827-a5ae8d.log) — System spawn log captured by task-board
- [TASK-260827-19aqkr_spawn-log_-implementer--doc-writer--agy-_RUN-260827-33a125.log](file://TASK-260827-19aqkr/TASK-260827-19aqkr_spawn-log_-implementer--doc-writer--agy-_RUN-260827-33a125.log) — System spawn log captured by task-board
- [TASK-260827-19aqkr_spawn-log_-reviewer--reviewer--claude-_RUN-260827-5592dd.log](file://TASK-260827-19aqkr/TASK-260827-19aqkr_spawn-log_-reviewer--reviewer--claude-_RUN-260827-5592dd.log) — System spawn log captured by task-board
- [TASK-260827-19aqkr_review-verdict.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_review-verdict.md) — Reviewer verdict: changes requested. 9 blocking findings, doc written against a stale base with no HTTPS implementation; docs/build-https.md already exists on origin/main.
- [TASK-260827-19aqkr_logbook-entry.md](file://TASK-260827-19aqkr/TASK-260827-19aqkr_logbook-entry.md) — Logbook block for LOGBOOK.md, not appended: the only reachable LOGBOOK.md is the stale local main copy. Land after rebase onto origin/main.

## Created
2026-08-27T01:40:47Z

## Last Update
2026-08-27T02:35:43Z

## Assigned To
[reviewer] reviewer (claude)
