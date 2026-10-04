# BUG-261004-16a407 \u2014 trust-pin-overrides-strict-findings: review verdict

Verdict: changes_requested. Route: to-dev. No implementation defect found in the reviewed pin-policy delta; producer recovery and current-candidate validation are required before acceptance.

## Blocking review findings

R1 \u2014 No accept-ready Change Request. `task-board worktree status STORY-261004-1aqcha --json` returned `change_requests: []`; the task change-request activity query also returned no events. This reviewer was supplied no Change Request revision. Consequently there is no revision to accept with accept_cr. Do not invent a revision or use reviewer commit_ack/done. Producer must publish the existing work through its tracked handoff, then route a revision-bound reviewer.

R2 \u2014 Hosted evidence is for an earlier full tree. Run 37170471503 validated cc3ba67211bc533558c3a6d0fe8bb6bce2ad7a2f. The three N5 files match that snapshot exactly (git diff --exit-code returned 0), but current HEAD is 7347499843dc7f0d2bc80f025d5eba4342d8060f and the current full source differs in internal/buildrepo/admission.go, internal/buildrepo/snapshot_budget_test.go, docs/repository-admission-limits.md and CHANGELOG.md. These are upstream changes, not unauthorized N5 edits. The admission/test delta is 383 insertions and 10 deletions. Latest main run 37180000622 is green but its audit code still contains the old pin bypass, so that run does not prove the combined candidate. Publish and host-validate the actual current candidate before the next review. Do not rerun locally on the stalled host.

## Independently inspected hosted evidence

https://github.com/relux-works/curator/actions/runs/37170471503

Downloaded artifacts with gh run download and read every lane's go-test.json. Independently counted the following on EACH of Test Linux, Test macOS, Test Windows, Race Linux, Race macOS: 52/52 TestGatePinPolicyFreshAndCached leaf cases, 2/2 TestDraftAuditPinDoesNotWaiveStrictFindings cases, 8/8 TestGateEmitsScriptAuditLabels leaf cases; audit package pass; zero failing Go events. Counting script exited 0. Hosted logs explicitly report go test exit=0 and platform-case gate exit=0 for all five lanes. 11/11 required jobs succeeded, including lint and interop. Two optional jobs were skipped; no claim is made for them. Workflow pins curator-spec rc.14 commit 43bf0a2506d5c354a73bbc3ea4623d4653db10c7. This proves the earlier snapshot, not the untested combined tree.

No local go build, go test, go vet or lint was run by this reviewer. Prior producer local red/mutant claims were not independently executed. Static analysis replaces execution as required by HOSTED-EVIDENCE MODE. git diff --check exited 0.

## Swept surfaces

| Surface | Result and bound |
| --- | --- |
| Strict pinned findings | Held by inspection: shared decideWithPins delegates to Decide, which blocks verifiable severity >= fail_on. Schema 3 at threshold, schema 4 above, and legacy schema 2 finding cases cover the policy. |
| Legacy pin requirement | Held: strict unpinned schema <3 still requires pin; clean pinned schema 1/2 allows. |
| Fresh/cache parity | Held: both auditSubjectWithOpaquePaths branches call the same function. Tests warm through Gate, check actual cache presence, then change policy and pin state. |
| Production reachability | Held: Gate/GateReadOnly are called from install.Project at internal/install/install.go:583/585, global install at global.go:198/200 and CLI at cmd/curator/main.go:1963/1965/2265. Install regression calls Project, checks explicit finding refusal and absence of source-audit binding. Unexpected directory read failures are fatal, not treated as absence. |
| Revocation | Held: checked before cache/fresh decision; strict and advisory pinned revocation rows block. |
| Advisory and thresholds | Held: pinned/unpinned findings warn in advisory; strict below-threshold and fail_on=off warn. Prior pinned silence was the defect, not an advisory behavior to preserve. |
| rc.14 warning controls | Held on hosted snapshot: 8/8 label rows per lane and full audit package pass; unchanged label generation remains separate from findings. |
| Spec clarification | Held: attached BUG-261004-16a407_spec-clarification.md contains the requested sentence for profiles/manager.md section 7. Latest brief explicitly requires a draft only; spec repository untouched. |
| Scope | Held: working delta contains only audit.go, audit_test.go and draftaudit_test.go. No N5 LOGBOOK.md/CHANGELOG.md edits or unrelated code. Existing board logbook note records producer findings. |
| Candidate validation/CR | Not held: R1 and R2 above. |

## Static red-first and mutation analysis

Base HEAD returns allow unconditionally for pinned content after the legacy check. The new audit matrix would fail 28/52 rows on that logic: seven pinned finding scenarios across two entries and two cache states expect block or warn, but receive silence. The remaining 24 control rows continue to pass. The updated install regression would fail both fresh/cached cases because old logic admits the finding and can establish a source-audit binding. Assertions name the network finding, avoiding unrelated failure as a false positive.

Required early-pin-allow mutant: killed by those 28 audit cases and both install cases. Required fix-only-one-path mutants: retaining bypass only in cache is killed by 14 cached audit cases; retaining it only in fresh is killed by 14 fresh cases. Production entry coverage prevents a helper-only fix from passing.

Reviewer-added narrowing mutant: change the strict finding comparison from >= to >. The high finding at fail_on=high would become warn, violating strict-schema3-at-threshold and strict-schema2-with-finding assertions on both entries and cache states (8 audit cases), plus both high-threshold install cases. This tests the exact boundary rather than deleting the gate. All kill statements are reasoned predictions, not claimed executions.

Free hunt: no additional code findings after sweeping policy precedence, cache policy recomputation, production call sites, source-audit absence handling, warning controls and scope. Coverage numbers describe this matrix, not all possible audit detectors or all schema versions.

## Next producer/reviewer cycle

Preserve the existing fix and tests. Obtain hosted green for the actual combined candidate, publish a valid tracked Change Request with its validation evidence, and hand that exact revision to a reviewer. No human decision or external blocker is identified; this is recoverable delivery rework. The run goal query returned no active goal. This artifact is also the review logbook record; LOGBOOK.md is intentionally untouched.
