## Status
done

## Review
light

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- TASK-260823-1wvgw8

## Blocks
- (none)

## Checklist
- [x] Deleting a schema-8 family from the conformance root makes the suite fail (consumption proven)
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Declare every schema-8 family in root-artifacts.tsv so a missing one defers its package
- [x] Add a conformance consumer for vectors/script-host-execution-policy.json
- [x] Name every schema-8 consumption case in platform-cases.tsv on all three runners
- [x] Prove family removal fails the candidate lane and a vanished case fails the ledger
- [ ] Dispatch candidate-conformance for 6001dc3 with the new coverage and attach the 3-OS matrix
- [ ] Route the evidence to the epic and to the skill-project-management auto-return pointer
- [ ] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260824-8dedd8, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260824-8dedd8)
Implementation on branch task/TASK-260823-2u5xov-schema8-consumption, commit b902023, PR #37 (open, not merged).

WHAT WAS ACTUALLY MISSING
- root-artifacts.tsv declared schema-cases/agent-skill-v8 but not csk-skill-v8, and nothing declared vectors/script-host-execution-policy.json. No Go test read the script-worker family at all.
- platform-cases.tsv carried NO row for any schema-8 consumer, so TestReleasedSchemaCases, TestReadAuthoritativeMarkerV4SchemaCases, TestModuleRootVectors and TestModuleRootVectorsDriveTheWholeBuild could each have been renamed or deleted with every lane still green.

CHANGE
- root-artifacts.tsv: csk-skill-v8 added to internal/skillspec; new internal/scriptpolicy row for the script family.
- platform-cases.tsv: eight rows naming the schema-8 consumption cases, required on all three runners, each tolerating exactly the skip class its case prints without a root (root-unset for the four deferred packages, root-content for the godriver case that guards its own read). root-unset is deferred-only, so the candidate lane leaves no tolerated skip.
- internal/scriptpolicy/conformance_test.go: closed policy identity and interpreter set checked against the suite bytes; every opt_in_case decides manifest acceptance and enforced/declared-only; every enforced shape refused with script_execution_policy_unsupported before any worker surface; the file top-level section set asserted in both directions against a classification table.
- Candidate-lane prose generalised from schema v6 in ci.yml, candidate-suite.sh, suite-plan.sh, root-artifacts.tsv. No check changed.

DECLARED GAP, NOT IMPLEMENTED HERE
audit_label_cases requires script-command-declared-only and script-command-unfiltered-declared-network in the audit record (manager profile 7, core 4.1.1). curator emits neither, and script-command-declared-only is NOT worker-dependent: it applies to every declared-only script command curator already installs. Classified by name in scriptHostExecutionPolicySections with owner STORY-260822-2h0v9j so it cannot read as covered. Needs its own task -- it changes audit decision semantics (always warn, never subject to fail_on, never reportable as an applied control), which is out of scope for a coverage commit.

conformance-claim-v5 deliberately NOT declared: curator publishes no conformance claim and has never consumed conformance-claim-v1..v4. A forced consumer would be a fake gate.

LOCAL EVIDENCE (all attached)
- family-removal: suite-plan.sh exit 1 under CI_REQUIRE_FULL_ROOT=1 for each of agent-skill-v8, csk-skill-v8, install-marker-v4, module-roots.json, script-host-execution-policy.json removed from the 6001dc3 root.
- vanished-case: platform-case gate FAILED by name when TestScriptExecutionOptInCases stopped running.
- candidate lane: all eight rows observed passing against the 6001dc3 root on darwin.
- default lane: against the materialised SPEC_PIN 00b1688a, every new row tolerated by its declared class, gate ok.
- gate-selftest 81/81 pass, ledger-consistency 80 rows ok across linux/darwin/windows, no-broad-suppression ok, golangci-lint ./... 0 issues, go build ./... exit 0.
agent completed: [implementer] developer (claude) (exit=124)
spawn run completed: claude (run=RUN-260824-8dedd8, pid=33210, exit=124)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260824-f63fa0, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260824-f63fa0)
REVIEWER SCOPE: verify only the PR 37 delta (a3abcf3) — consumption assertions and the family-removal negative proof — plus the qualification evidence resource (run 32689488293 green with candidate 6001dc3). Targeted tests; cite green lanes; do not run the full suite locally.
REVIEW VERDICT: CHANGES REQUESTED (to-dev). Code delta accepted, routing not done. See TASK-260823-2u5xov_review-verdict.md.

VERIFIED INDEPENDENTLY (not taken from producer evidence)
- Consumption is real: 8 mutations of the 6001dc3 script-host-execution-policy family (new section, deleted section, renamed policy, widened interpreter set, flipped accepted x2, flipped mode, emptied opt_in_cases) each fail the new internal/scriptpolicy cases with the right message. The surface driven is production (skillspec.Load, ScriptExecutionPolicy, ScriptInterpreters, scriptpolicy.Enforced/Admit/Code).
- Family removal: reproduced all five. CI_REQUIRE_FULL_ROOT=1 suite-plan.sh exits 0 on the full 6001dc3 root and exits 1 naming the deferred package for each of script-host-execution-policy.json, csk-skill-v8, agent-skill-v8, module-roots.json, install-marker-v4.
- Tolerance surface has no hole: root-unset is deferred-only, so the 7 rows carrying it are required-in-fact with a serving root; the one root-content row (godriver) is already covered by internal/moduleroots artefact row.
- Qualification: run 32689488293, workflow_dispatch on main at a3abcf3 (WITH the delta), 14/14 jobs SUCCESS. Job logs confirm CANDIDATE_REF 6001dc33281b94a4ec7442ab15278550dd0f51d9, manifest sha256 803918bf..., CI_REQUIRE_FULL_ROOT=1, served=42 deferred=0. All 8 new ledger rows observed by the platform-case gate on all three runners (godriver row excl on linux, as its package is excluded there by the root qualification vector).
- Merge tree a3abcf3 == PR head a0e1557 (f7f4588e...), so the merged tree is what was reviewed.
- Declared gaps are honest, not forced fits: audit_label_cases named with owner STORY-260822-2h0v9j; conformance-claim-v5 correctly not declared.

UNMET AC CLAUSE: evidence routed
- EPIC-260822-18ylpq notes carry no qualification evidence (no run id, no candidate identity, no matrix).
- STORY-260822-2lvw0e checklist item 1 (unblock skill-project-management TASK-260822-hje0ya, branch task/go-v1-switch) unchecked; nothing on the board references the auto-return pointer.
- Task checklist items 6, 12, 13 unchecked. Board-wide grep for 32689488293/6001dc3 finds the evidence only in this task resources.
Cause: producer run exited 124 post-PR.

NEXT CYCLE (board-only, no code): 1) note the qualification on the epic; 2) route the unblock to TASK-260822-hje0ya and tick story item 1; 3) attach the 3-OS matrix as its own task-scoped artifact in the TASK-260822-c0rxj7 *_green-matrix.md shape; 4) tick items 6/12/13, back to to-review. Do NOT re-review the code delta -- verify routing only.

NIT (non-blocking): this delta is the first user of the root-unset skip class in platform-cases.tsv (0 rows before, 7 now) but that file header still lists only platform-control/host-capability/root-content/opt-in. Gate unaffected (class lives in skip-classes.tsv); fold into the next edit of that header.

No commit_ack from this reviewer run; PR 37 is already merged as a3abcf3.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260824-f63fa0, pid=97281, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260823-2u5xov_spawn-log_-implementer--developer--claude-_RUN-260824-8dedd8.log](file://TASK-260823-2u5xov/TASK-260823-2u5xov_spawn-log_-implementer--developer--claude-_RUN-260824-8dedd8.log) — System spawn log captured by task-board
- [TASK-260823-2u5xov_family-removal-suite-plan.txt](file://TASK-260823-2u5xov/TASK-260823-2u5xov_family-removal-suite-plan.txt) — suite-plan.sh exit 1 under CI_REQUIRE_FULL_ROOT=1 for each of the five schema-8 artefacts removed from the 6001dc3 root
- [TASK-260823-2u5xov_vanished-case-platform-cases.txt](file://TASK-260823-2u5xov/TASK-260823-2u5xov_vanished-case-platform-cases.txt) — platform-case gate fails by name when a schema-8 consumption case stops running
- [TASK-260823-2u5xov_candidate-lane-platform-cases.txt](file://TASK-260823-2u5xov/TASK-260823-2u5xov_candidate-lane-platform-cases.txt) — all eight schema-8 consumption cases observed passing against the 6001dc3 root (darwin)
- [TASK-260823-2u5xov_default-lane-platform-cases.txt](file://TASK-260823-2u5xov/TASK-260823-2u5xov_default-lane-platform-cases.txt) — default lane against the committed SPEC_PIN: every new row tolerated by its declared skip class
- [TASK-260823-2u5xov_qualification-evidence.md](file://TASK-260823-2u5xov/TASK-260823-2u5xov_qualification-evidence.md)
- [TASK-260823-2u5xov_spawn-log_-reviewer--reviewer--claude-_RUN-260824-f63fa0.log](file://TASK-260823-2u5xov/TASK-260823-2u5xov_spawn-log_-reviewer--reviewer--claude-_RUN-260824-f63fa0.log) — System spawn log captured by task-board
- [TASK-260823-2u5xov_review-verdict.md](file://TASK-260823-2u5xov/TASK-260823-2u5xov_review-verdict.md) — Reviewer verdict: changes requested (evidence routing unmet); consumption assertions and family-removal proof independently reproduced and mutation-tested; run 32689488293 verified green on 3 OSes with candidate 6001dc3

## Created
2026-08-23T19:42:43Z

## Last Update
2026-08-24T09:24:58Z

## Assigned To
[reviewer] reviewer (claude)
