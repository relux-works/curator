# TASK-261013-g96cen — wave6-cips-review-findings — reviewer A

Record-only review of CR-TASK-261013-g96cen-1 revision 1.
Base: a01542857678108a77d5e1b1f051989b00384f1d.
Candidate tree: d71ef82972d8c5d40d681e26f13c974faec55ce6.

No accept_cr, reject_cr, status mutation, repository edit, commit, test, build, renderer or LOGBOOK write was performed. The explicit reviewer-A hard rule overrides the generic lifecycle instructions. `spawn goal` reports this run is not goal-bound. Reviewer B owns the decision.

## Findings

```yaml
findings:
  - id: A1
    severity: robustness
    severity_reason: "P2 — two updated summary figures still prescribe fresh materialization for every round, contradicting the required same-workspace refusal repair and risking loss of valid edits in an implementation following those summaries. Documentation defect, not a demonstrated runtime failure."
    locations:
      - "cips/diagrams/donor-layout.puml:84"
      - "cips/diagrams/donor-lifecycle.puml:88"
    evidence: "The layout says 'fresh materialization each round'; the lifecycle says 'R-TRANSFER each round: fresh B/C materialization'. CIP-0009 D5a's repair paragraph (line 205) instead requires the original caller to admit repair on the same workspace and preserve valid edits. The detailed transfer figure correctly reopens that workspace. Adding a link and a repair note below the unconditional summary does not remove the contradictory copy."
    requested_change: "Limit fresh materialization to initial work and candidate-based review rework; explicitly exempt the separately admitted bounded repair round, which retains its existing workspace. Update the gaps report's figure-consistency claim accordingly."
  - id: A2
    severity: note
    severity_reason: "P3 — make the comparison domain explicit; the intended meaning may already be inferred from 're-observation', but an explicit condition avoids classifying a successful repair as contradictory evidence."
    locations:
      - "cips/CIP-0009-donor-side-deployment-and-bridge.md:203"
      - "cips/diagrams/remote-worker-transfer.puml:64"
    evidence: "The accepted finding limits recorded/refused or recorded-digest contradiction to two observations of the same quiesced tree. The new prose and figure retain the compared values but omit that condition. Nearby repair/re-export also uses new linked effects, legitimately changing refused to recorded after editing the tree."
    requested_change: "State that contradiction comparison applies to the same unchanged quiesced tree under the same pinned export inputs, not across an admitted repair that changes the tree. Keep terminal replay and linked re-observation rules unchanged."
```

## Swept surfaces and requirement coverage

| Surface | Review result |
|---|---|
| Naming namespace | Disposition replaces launcher audit-token gate references with Decision 0021/launcher §10 ownership; R-CS2 is historical, not a rule. Decision 0005 retitle preserves pathname and explicit old anchor. Decisions 0010/0013/0019 qualify composition references. No unrelated behavior change found. |
| Wire/status register | Owner short transports, qualified domains and helper kinds retained; result DRAFT fields distinguished from transport/lifecycle. Canonical index stays external. Unregistered onboard proposal confers no authority; no replacement action allocated. |
| Muse | CIP-0010 question 4 and disposition consistently separate decided priority, proposed channel and open qualification; API-key lane remains distinct. |
| Invocation closure | CIP-0009 D5c, future qualification cases, protocol and turn figure cover transitive interpreted inputs and direct/indirect heavy-form bypass, with inert positive controls and R-VERIFY fallback. |
| Base binding | D5a steps 1/5 and protocol require immutable-B manifest derivation, archive path/kind/mode/content verification, missing/extra refusal and independent repository-B import recomputation with existing base_commit/manifest_digest checks. |
| Mode semantics | D5a/protocol and detailed figures cover mode-only changes, equal-blob type swaps and deletion prior mode. No result schema or code file changed. |
| Refusal and repair | Four outcomes, retained immutable refusal, new effect identities, current-authority plan-bounded repair, permission/infrastructure distinction and per-act waiver are present. Standing discard stays off. A1/A2 concern summary consistency and comparison precision. |
| Cleanup | Original-admission bounded expiry consent, separate physical execution evidence, holds, pre-retiring collection, all-blob digest/size durability outside deletion, current retrieval authority and authenticated release are explicit. No unknown-as-absent, duplicate launch after lost reply, automatic ALLSTOP recovery or cleanup-authorized worker repair introduced. |
| Figures | Read all changed PlantUML/DSL source deltas; 10 changed PlantUML sources plus C4 are linked in README. Detailed transfer/turn/retirement branches reflect custody and retention. Two stale unconditional summary sentences remain (A1). Rendering/semantic parser results unknown. |
| Gaps outcome | TASK-261013-g96cen_gaps.md covers 8/8 records and accurately names external publication and qualification work. Its broad figure-alignment claim needs the A1 correction. Reported validation was not represented as reviewer-run evidence. |
| Scope/hygiene | All 20 changed files are Markdown/PlantUML/DSL. Revision notes present. Diff is confined to naming/status/transfer/invocation and affected figures. Existing historical review identifiers were preserved as the brief requires; no mechanical UTF-8 corruption observed. |

## Evidence and limits

Independently read the exact base-to-candidate diff across all 20 paths, current affected source sections, the attached gaps outcome and figure index. `git diff --name-only d71ef829` produced no paths: the observed working files match the candidate tree. `git diff --check a015428 d71ef829` exited 0.

Read local immutable neighbouring sources: dispatcher 74687e6 §5.2.1 confirms DRAFT field semantics and closed refusal vocabulary; broker df8c738 §10.2 confirms Muse priority/proposal/qualification; helper 7b59da74 §7 confirms the earlier transfer baseline. Local wiki HEAD 26cf60e390c61a75a12391fbea182bc5637645e1 contains the typed shared-rule/wire index and matching naming-category rule. This is local source evidence, not proof of current hosted main or external URL/anchor reachability.

Read-only scan of all 225 added diff lines found zero specified private finding slugs, local /Users or /home paths, board/session links, and review seat/event patterns. These patterns are bounded checks, not an exhaustive classifier for arbitrary personal/host names; the delta was also manually reviewed. Existing review identifiers in unchanged or otherwise retained baseline prose are not newly introduced private findings.

No tests/builds or renderer runs. Producer reports schema/vector validation and a bounded source checker; these were read as producer evidence, not rerun or independently certified. Cross-module unique-token resolution, hosted freshness, external anchors and runtime/provider/OS qualification remain unmeasured. No acceptance or delivery verdict is issued by this artifact.
