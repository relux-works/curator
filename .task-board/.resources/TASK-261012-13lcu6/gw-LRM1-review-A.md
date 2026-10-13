# TASK-261012-13lcu6 — launcher-v030-muse-admission-and-failclosed-hosted

Reviewer A, revision 2, RECORD-ONLY. No acceptance/rejection or status mutation.
Base f08f566dd378ea27ce23ca0351ec793e95a696b6; candidate tree 46ad6f570db826f7980ae30f410cb34f77af94f7.

## Findings

- id: A2-01
  severity: note
  severity_reason: P2 — stale public documentation; the explicit runtime availability notice and fail-closed code prevent this from being a launch bypass.
  location: README front-door table and contract link; SPEC §10 preamble and §4.2 Pi paragraph; diagrams/components.puml, launch-paths.puml, launch-states.puml, artifact-layout.puml and diagrams/README.md.
  finding: CURRENT figure panels still depict receiver execution without the unavailable notice, and retain SPEC 0.5.0-draft. README's front-door table likewise calls hosted an existing contract without the nearby availability qualification. SPEC §10 says the executable specification remains 0.5.0-draft, while the actual version is 0.5.1-draft; §4.2 retains a stale v0.5.48 Pi pin statement.
  action: Follow-up documentation correction. Revision-2 instructions explicitly prohibit redrawing figures and require recording the discrepancy; the gaps artifact does that. No code rework requested for this observation.

- id: A2-02
  severity: note
  severity_reason: P2 — evidence placement/completeness gap, with the actual candidate identity and successful validation established separately.
  location: TASK-261012-13lcu6_change-request_rev2-validation.log.
  finding: This log records the successful handoff gate and coverage bound, but not the required new base, per-file apply results, SPEC conflict resolution or 23-file delta-identity check. Those details are in the gaps artifact (apply outcome summarized as 24 clean and one SPEC conflict) and producer evidence instead.
  action: Preserve an explicit cross-reference or consolidated validation record for reviewer B; do not describe the CR validation log alone as containing all five required items.

- id: A2-03
  severity: note
  severity_reason: P2 — bounded review coverage, not evidence of a code defect.
  location: review source set.
  finding: gw5-decisions.md, gw4b-decisions.md and gw4-decisions.md were not present in the task, parent Story or parent Epic resource listings inspected. Neighbouring module specification sections were not independently revalidated. The wave-5 comparison below is against the supplied brief and preserved repository text; it is not certification against those missing source documents.
  action: Reviewer B should reconcile against the original decision resources and neighbouring specifications before recording the deciding verdict.

## Swept surfaces and evidence

| Surface | Review result |
|---|---|
| Pin/API drift — original brief §1, SPEC §§4.2/4.4/7 | go.mod/go.sum pin v0.5.62 with no new replace. Production resolvePermission preserves the probed ToolRelease into the builder request; only the exact Muse help-evidence-required error defers mapping to real BuildLaunch admission. The module-attested diagnostic follows admission. No new wire member/refusal code introduced by this delta. |
| Refusal meaning | Muse malformed/empty/inconsistent version rows replace obsolete release-table refusal; positive novel-build admission reflects the explicit new policy. Both postures retain release-claim mismatch rejection. Empty, example-only and prefix-only yolo help evidence refuse. Tests exercise run and the real builder, and assert no child. Baseline parity permits exactly two Muse lineup substitutions, each once, rather than dropping parity assertions. |
| Hosted — original brief §2, SPEC §§3/4.8/4.9/6 | run → launch → runHosted → Transport.Run uses unavailable LookupReceiver in production. It refuses before receiver execution with existing session_host_missing; machine defaults keep the earlier session_host_default_not_ready gate. Eight added entry-point rows cover four selectors × two postures and no receiver/provider start. The production transport test also checks terminal cleanup. M-H22 narrows refusal to allow a PATH receiver and is required by mutantguard. |
| Native/network preservation | No new native execution route or network-policy mutation. Native/explicit native/untracked behavior stays in the existing routing; tests retain native-zero-receiver lookup coverage. Muse network tuples remain outside the documented child-scope ceiling. |
| Release text — original brief §§3/4 | CHANGELOG has 0.3.0 (pending tag), Muse admission/alias/fragment v3, already-landed direct network and hosted-unavailable notice. SPEC header and §8 are 0.5.1-draft; revision notes keep 0.5.1 first and LR6 0.5.0 second. Stale-copy exceptions listed above. |
| Wave-5 semantics — SPEC §§10.1–10.6 | Preserved target text defines one data-export path with whole-export refusal, exact-candidate review, no agent-account builds/tests, provider-only trusted harness access, workplace zero egress, trusted-only standing loops, qualified per-OS capabilities rather than OS exclusion, narrowing-only controls and explicit recovery. Unpublished fields/formats remain publication questions. This is target prose, not implementation proof. |
| Figures | Four unchanged LR6 figures remain linked from SPEC and retain target publication gates. Hosted CURRENT-panel discrepancy truthfully recorded in gaps, as explicitly required by revision-2 brief. |
| Public hygiene | Inspected candidate delta adds no real local absolute path, personal name, session link or board-resource URI. Gate snapshot message removes the workspace path. This is a changed-content check, not certification of every historical repository file. |
| Gaps artifact | Correctly records deferred hosted activation, unsupported Muse network tuples, unavailable bug disposition, diagram/version discrepancies and pending tag. Gate referenced there is the earlier successful same-tree run; the handoff log records a later successful same-tree run. |

## Revision and validation evidence

Independently compared rev1 patch index blobs with revision-2 base/candidate blobs: 23/23 non-README/SPEC paths have identical pre-images AND post-images. No code was edited for review. git diff --check on the exact base/candidate delta returned 0.

Both gate commits resolve locally to candidate tree 46ad6f570db826f7980ae30f410cb34f77af94f7:
- a15a715f95d7d5744fdb9e23df6b19ac76c7919c: attached hosted_gate.json, run 38363575922, four required Linux/macOS test/race jobs successful; optional self-hosted job skipped.
- 19a36991606be51b7c445a9c40c3dbda4d7148be: CR revision-2 validation log, run 38364602774, same four required jobs successful, command exit 0. Log reports exact_command_shard required=1 green=1 failed=0 missing=0; test_case_coverage=unknown.

These are accepted attached hosted results, not tests rerun by reviewer A. No local go test, go build, go vet or suite was run in this review. Producer reports no local builds/tests; no contrary execution was established from the inspected evidence, but this is not an exhaustive historical process audit. No LOGBOOK.md write, commit, tag or code mutation.

## Record-only conclusion

No P0/P1 implementation defect established on the reviewed surfaces. Three P2 notes above, including the explicit source-coverage limit. Reviewer B owns reconciliation and the deciding verdict. The run is not goal-bound (spawn goal returned none). Status and CR state deliberately left unchanged under the latest reviewer-A hard rule.
