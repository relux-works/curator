# TASK-261011-1pvp8b — R-LR2 diagrams: Reviewer A, record-only

Reviewed CR-TASK-261011-1pvp8b-1 revision 1, base `1d60804709d57d81710a5381e664125307d0d31e`, candidate tree `252c43422e9f03557f701a27086ee0c82e98d3b2`.

Recommendation to Reviewer B: correct the two state-diagram findings below before acceptance. This is an advisory review record, not a CR verdict. No accept_cr, reject_cr, status mutation, checklist mutation, commit, repository edit or LOGBOOK edit was performed. The explicit record-only brief overrides the generic first/status/verdict instructions. `spawn goal` reported that this run is not goal-bound; the directive checkpoint reported no directives.

## Findings

### A-F1 — Do not mark registered one-shot identity revocation optional
- severity: regression
- severity_reason: P2 — the new diagram weakens an explicit owner cleanup decision, although this docs-only candidate does not change runtime behavior.
- location: `diagrams/launch-states.puml:75` (also terminal transition at line 99).
- evidence: `Retired` is labelled `Account retired / optional separate registered one-shot revocation`. The owner precondition `owner-decisions-20261010-audit.md`, D-IDENTITY, says the dispatcher revokes identities registered as one-shot as a separate operation after account retirement; standing identity revocation remains operator-only. It does not make that one-shot operation optional.
- nuance: existing SPEC §10.4 says D-IDENTITY “permits” revocation only for registered one-shot identities. That defines its authorization boundary; it does not establish a new optional-cleanup decision. The added diagram should not resolve that wording by explicitly making the operation optional.
- requested correction: remove `optional`; show the separately owned post-retirement operation for registered one-shot identities while preserving the standing-identity exclusion. Do not invent an acknowledgement schema or broaden revocation authority.

### A-F2 — Distinguish pre-execution refusal from runtime failure
- severity: robustness
- severity_reason: P2 — the new state map collapses outcomes that SPEC §10.5 expressly requires the executor to distinguish.
- location: `diagrams/launch-states.puml:89` and lines 68–74, 98.
- evidence: `Harness --> TargetRefusal : recheck / renewal or authentication failure` combines a pre-execution binding recheck failure with failures after harness execution. The same TargetRefusal node receives intake and privilege failures and terminates the invocation. SPEC §10.3 says renewal timeout fails the turn and Claude authentication failure requires re-enrollment/restart; §10.5 separately requires pre-execution refusal, failed launch, harness exit, signal termination and unknown evidence to remain distinguishable. §10.4 leaves their wire transport unpublished, not their semantic distinction.
- requested correction: keep the pre-exec recheck refusal separate; represent runtime renewal/authentication failure through execution outcome/failure evidence and preserve dispatcher stop/absence/cleanup obligations. Label unpublished evidence transport without assigning numeric exits or inventing a surviving executor process.

## Swept surfaces

| Surface | Review result |
|---|---|
| M-LR1 `components.puml` | Current caller-side processes, composer as an in-process library, direct executor, ax and hosted receiver are separated from the target root transition, agent executor, broker and supervisor. Refusals, persistent owners and unspecified service identities are visible. No additional actionable finding. |
| M-LR1 `launch-paths.puml` | Resolve/build/provider-limit/composition flow, direct versus ax versus hosted handling, FD4 classification, no fallback, target receipt-before-drop, binding recheck, stop and unknown-evidence branches checked against SPEC §§4, 4.6, 4.8–4.9, 6 and 10. No additional actionable finding. |
| M-LR2 `launch-states.puml` | Current invocation states remain transient, target state schema is explicitly illustrative, and unknown evidence never releases capacity. A-F1/A-F2 need correction. |
| M-LR2 `artifact-layout.puml` | Config/fragment/home/payload ownership, hosted FD3/FD4 versus target FD3, plan bounds, broker material versus metadata, root start receipt and dispatcher records checked against SPEC §§4.1, 4.4b–4.7, 4.9, 10. No additional actionable finding. |
| Phase and owner decisions | Current interactive and hosted Phase 1 versus unimplemented local-headless v0 are explicit; v1 daemon and SH2 activation stay later/open. D-EXECUTOR, D-PATH, optional board dependency, donor deferral and later credential lanes are preserved. D-IDENTITY exception recorded above. No invented wire schema, numeric executor exit, activation or credential qualification observed. |
| Links/index/outcome | All 4/4 audit filenames present; 4/4 linked from README and SPEC; 4/4 contain defining SPEC links; index maps all four to B4 rows and sections and records unpublished boundaries. Attached producer index is byte-identical to candidate index (`cmp`, exit 0). Structural index claims hold; semantic completeness is subject to the findings above. |
| Scope/hygiene | Exact CR delta: 7/7 paths are documentation, with SPEC/README changes limited to diagram links and descriptions. No product code, test, build or LOGBOOK edits. No first names, host-local absolute paths or session links observed in added content. |

## Evidence and limits

Sources: candidate SPEC/README and all four figure sources; audit at `48789217` B4 M-LR1/M-LR2 and D1 R-LR2; contract table at `177cb554`; owner-decision precondition with D-GOALS, D-IDENTITY and the 2026-10-11 amendments. The owner decisions and SPEC §10 supersede the table's earlier open executor/PATH/naming choices only as stated.

Read-only inspection performed here: exact base-to-candidate diff/stat; source-to-spec review; all seven working-file Git blob hashes compared to candidate tree (7/7 match); producer outcome comparison. No tests, builds, lint or renderer were run, and no prior gate pass is adopted as fresh evidence. PlantUML rendering is unverified. These coverage counts describe document inspection, not runtime gate coverage or capability proof.

Reviewer B retains the decision and lifecycle authority. These are ordinary documentation corrections, not an external blocker or a new human-only decision.
