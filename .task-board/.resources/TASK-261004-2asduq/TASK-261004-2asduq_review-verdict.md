# Review verdict: changes_requested

Task: TASK-261004-2asduq — launch-command-environment-fragment-design.
Reviewer run: RUN-261004-c2c90f. Review date: 2026-10-05.
Route: analysis, for researcher handoff repair and another reviewer cycle. No human decision or product implementation is required for this repair.

## Finding R1: missing reviewable Change Request

The research content passes this review, but the required acceptance transaction has no candidate revision. `task-board --json worktree status STORY-261004-1ffwh8` (design-launch-project-views) returned `change_requests: []`. This assignment contains no Change Request Under Review revision. The task's change-request activity query returned no events; both research outcome attachments are present. The reviewer run is not goal-bound (`task-board spawn goal` returned that result).

The companion evidence lines 209–211 says the files must be attached before the researcher handoff and names `task-board handoff TASK-261004-2asduq --role researcher` as the final command. Attachments exist, but no published Change Request is available for the binding review instruction's `accept_cr`. These observations establish the missing CR, not why it is missing or whether a previous handoff command was attempted.

Required repair: preserve the two research artifacts, have the tracked researcher publish the candidate through the supported handoff flow, and route a reviewer bound to that revision. If the handoff returns no CR, resolve that recoverable workflow mismatch before another acceptance attempt. Do not mark this leaf done, manufacture a revision number, or use commit_ack. No content rewrite is requested by this review. Design acceptance and scheduling remain separate operator decisions.

## Swept surfaces

| Surface | Result |
|---|---|
| Brief and template | Held: all template sections, three real options, recommendation, concrete modes/defaults and seven decision-ready questions. |
| Five-environment inventory | Held: Claude, Codex, OpenCode, Pi and Muse; documented/measured/unknown bounds are distinguished. |
| Admission and precedence | Held: protected snapshots, per-category/per-item approval, selected-profile fallback, whole-skill replacement, explicit MCP conflicts and bounded permission grants. |
| Hostile repository | Held as proposed design: automatic discovery exclusions, transitive references, TOCTOU, replay, secret references, same-UID limits and unsupported-adapter refusal. No containment proof is claimed. |
| Commands and tracking | Held: one protected dispatcher, fixed append, earlier-shim conflicts, provider refusal set and destination-side PATH identity; no tracked capability before transport exists. |
| Homes, leases and login | Held: stable per-project/profile homes, busy-view refusal, resume/GC ownership, login cost and unresolved credential mechanism. |
| Migration, leaves and tests | Held for research scope: legacy/enrolled distinction, spec changes, ordered future leaves and meaningful negative/mutant test plan. No execution authorized. |
| Evidence | Held within stated bounds: ten source/spec spots checked below; native probes accepted as producer-reported evidence, not independently rerun. |
| Public safety and diff | Held within inspection bounds: only two untracked CIP research files; tracked diff empty; LOGBOOK.md unchanged; no personal filesystem paths or trailing whitespace found. Manual review found no credential values or internal hostnames in the outcomes. Historical account access cannot be independently proven from prose; the producer records synthetic fixtures and no credential/login operations. This review performed none. |
| Lifecycle | Not held: missing Change Request revision (R1). |

No additional blocking content findings in the swept surfaces. Free hunt: no additional findings.

## Independent citation spot-checks: 10/10 supported

Read with `git show <pin>:<path>` and numbered source lines; all source-read commands exited 0. Curator pin ca1b776 is an ancestor of the current local main (check exited 0); the draft explicitly dates and pins its baseline rather than claiming current remote freshness.

| Repository/pin | Locator | Verified claim |
|---|---|---|
| curator ca1b776 | internal/envprofile/managed.go:63–85 | Profile/environment home key; LaunchDir separate. |
| curator ca1b776 | internal/envfragment/envfragment.go:23–70 | v2 default, Muse v3, emitted structure lacks command environment. |
| curator ca1b776 | internal/envprofile/managed.go:2290–2312 | MCP fragment member conditional on materialized surface. |
| curator-agent-launcher d092035 | internal/composition/composition.go:74–125 | MCP channel flags conditional on non-nil MCP; owned literals/name lookup composition. |
| curator-agent-launcher d092035 | internal/mapping/mapping.go:18–32 | Only Claude, Codex and Pi provider mappings. |
| curator-agent-launcher d092035 | internal/fragment/fragment.go:324–333 | Parser admits v1/v2 only. |
| curator-agent-launcher d092035 | internal/fragment/fragment.go:493–506 | Reserved prepend and two-level home-root assumption. |
| curator-spec 43bf0a2, rc.14 | protocol/environments.md:2583–2606 | Singleton machine-current command limitation, reserved names, separate hybrid scope. |
| curator-spec 43bf0a2, rc.14 | protocol/environments.md:3155–3210 | Fragment revisions, closed fields, strict-MCP channel. |
| curator-spec 43bf0a2, rc.14 | decisions/0017-environment-credential-modes.md:113–136 | Credential sharing is not sandboxing; macOS Claude shared mode remains refused. |

## Validation bounds

Independently reran read-only Git/source checks and a two-file path/whitespace scan: both files have zero scanned personal-path matches and zero trailing-whitespace lines. `git diff --quiet HEAD --` and `git diff --quiet HEAD -- LOGBOOK.md` both exited 0. Worktree status showed only the two expected untracked research files. No product code, tests, configuration or LOGBOOK edits were made by this reviewer.

P1–P6 contain versions, synthetic fixture descriptions, native commands, exit statuses and explicitly limited observations sufficient to reconstruct their reported cases. Accepted these as attached producer evidence; did not rerun native binaries, login experiments, builds or Go tests under hosted-evidence mode. The producer's 13/13 and 14/14 document audits were not rerun as identical scripts and are not independently certified here. The proposed implementation test plan is not a green production suite.

The only requested change is R1, a recoverable producer handoff repair. The CIP remains a draft pending operator design acceptance, regardless of later research-review acceptance.
