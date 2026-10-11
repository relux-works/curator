# TASK-261011-23nfdx — R-CS3 diagrams — reviewer A, record only

Review target: CR-TASK-261011-23nfdx-1 revision 1, base f61feedffbddba8b9123cb9051547c5013da71cb, candidate tree cf9b9d78ef2fa80ca9ba2fc8b482b0a74a71ac1e. Recommendation to reviewer B: repair the four findings below before acceptance. This is not a CR verdict or lifecycle transition.

Authority: curator audit 48789217, B4 M-CIP1/M-CIP2 and D1 R-CS3; contract table 177cb554; owner-decision precondition retrieved from TASK-261011-2xdx9u, including the 2026-10-11 decisions. Read the R-CS2 second-review record and the producer's TASK-261011-23nfdx_figures.md through board resource get. The run goal query reports no bound goal.

## Findings

### A-F1 — Scope PATH transformation/refusal to the launch lane that requires it

- severity: regression
- severity_reason: P2 — the figures introduce refusals for supported context-only and enforced-script lanes.
- Evidence: cips/diagrams/configuration-publication.puml:29–38 always transports/applies the command transform and refuses the child for absent/excluded PATH. CIP-0002:157 expressly separates context admission from command transport so context-only opt-in does not require PATH modification; :188 refuses command capability, not every context-only launch. Similarly, tool-admission.puml:45–56 applies the admitted-base D-PATH transform and missing-PATH refusal before reaching enforced script-worker-v1, although CIP-0007 Design / Option B says that lane constructs PATH solely from the resolved interpreter and declared exec names and discards inherited PATH.
- Requested repair: show an optional command-capability branch in configuration-publication, retaining context-only launch without a command transform. In tool-admission, branch by execution lane before applying D-PATH; keep the declared-exec-only script path independent of the managed-harness base-PATH gate. Preserve shared admission checks that actually apply to both.

### A-F2 — Classify an existing hook cache before publishing replacement bytes

- severity: regression
- severity_reason: P2 — the sequence shows mutation before deciding whether an unknown/custom cache requires explicit regeneration.
- Evidence: configuration-publication.puml:81–86 sends canonical bytes and generation metadata into the persistent Hook cache before the alt for unknown/custom/unreadable/partial cache. CIP-0004 Compatibility and migration item 3 (:110) requires recognition and option preservation first; unknown/custom/unreadable caches require explicit regeneration with an explicit global choice. The current sequence has already replaced the cache when it reports that condition.
- Requested repair: read/classify first; put publication on the recognized-cache branch, or on a separately explicit regeneration branch after global choice. Keep reload as a subsequent independent action. Do not imply a committed install fails because a custom cache cannot be classified.

### A-F3 — Show the specified tool-publication transaction and failure recovery

- severity: robustness
- severity_reason: P2 — the requested recovery view omits an existing proposed contract and describes it as unspecified.
- Evidence: tool-admission.puml:41–43 goes directly from sealed-file provisioning/verification to lock/store publication and alias exposure, then says store/publication failure recovery details are unspecified and there is no transaction order to draw. CIP-0007 Specification changes item 8 (:489–494) explicitly requires operation-private download/verification before the manager-home transaction, journaled store/marker targets, protected-boundary/preimage revalidation, and mutable-target rollback distinct from immutable-store retention. Security considerations / TOCTOU and staging and the transaction test-plan row reinforce this sequence.
- Requested repair: depict that proposed staging → verified publication-under-lock/journal order and the failure/rollback branch. Keep genuinely unpublished encoding details marked unspecified; do not replace the already-described recovery contract with a generic gap note. This need not claim implementation or invent a journal format.

### A-F4 — Retain operator confirmation for orchestrator-driven donor admission

- severity: bypass
- severity_reason: P2 — the retained actor alias permits interpreting an onboard grant alone as the authority for admission/package issue.
- Evidence: donor-first-deployment.puml:11 labels OP as operator OR orchestrator with onboard grant. That same OP performs the admitted branch, pins/grants and package issuance without an operator-confirmation condition. CIP-0009 D4 step 5 (:151) requires an orchestrator with onboard grant PLUS operator confirmation. The figure's generic operator-admission note does not distinguish confirmation by the actual operator from action by the aliased orchestrator.
- Requested repair: qualify the orchestrator alternative with the required operator confirmation, preferably a visible admission guard/refusal before pinning and package issue. This is a retained inconsistency in an extended figure, not a claim that revision 1 introduced the old actor label. The brief expressly requires existing extended figures to match current prose.

## Swept surfaces

All 12 PlantUML sources and the retained C4 source were inspected as text. Coverage below means documentary inspection, not rendered or runtime correctness.

| Surface | Defining prose reviewed | Result/bound |
|---|---|---|
| donor-side.puml and donor-side.dsl | CIP-0009 D0–D5/D9/D12; remote-worker publication register | Variant/process roles, separate provisioner, disabled B, owner/phase notes and pending wire ownership retained; no additional blocking finding. C4 is explicitly the B target with A/C substitutions in notes. |
| donor-first-deployment.puml | CIP-0009 D0–D4/D9/D10/D12 | Trust refusal, same-key retry, provisioning rollback and state owners present; A-F4. |
| join-handshake.puml | CIP-0009 D4/D6/D9/D12 | DNSSEC versus declared invitation branch, no failure fallback, challenges/tombstones, package validation and CA distinction preserved. |
| remote-worker-turn.puml | CIP-0009 D0–D5/D12; CIP-0008 R3/R5/R9 | Caller-composed plan, variant boundaries, durable admission/cursors, signing/carrier separation and disconnect recovery represented. |
| remote-worker-retire.puml | CIP-0009 D7/D12 | Local lifecycle versus project/identity operations separated; process absence before release; offline outcome unknown; no donor application of one-shot revoke exception. |
| donor-variants.puml | CIP-0009 D0–D3; CIP-0008 qualification boundary | A limits, B separate provisioner/disabled status, C sandbox and key/network differences represented; later A availability stays open. |
| donor-lifecycle.puml | CIP-0009 D7/D12 | Independent axes, preserved donor identity, signed revoke and no offline-purge inference represented; descriptive rather than closed schema. |
| donor-layout.puml | CIP-0009 D0–D3/D5/D7/D12 | Variant roots, keys/state, services/socket and project CA/KRL ownership shown; unknown filenames/retention marked. |
| credential-source-lifecycle.puml | CIP-0003 migration/disposition; CIP-0010 C1–C6/source-policy revisions; CIP-0011 §§2–3 | Per-home/helper/node/broker distinctions, file-first backend, no fallback, binding/expiry and first executor represented; no runtime qualification inferred. |
| configuration-publication.puml | CIP-0002 admission/leases/command capability; CIP-0004 activation/cache migration; CIP-0006 owned keys/MCP | A-F1/A-F2; settings negative intent, per-home journal and release conflicts otherwise represented. |
| tool-admission.puml | CIP-0007 Option B, recommendation, security and specification changes | Revocation-before-selection, lock reauthorization and no C probes represented; A-F1/A-F3. |
| audit-transport.puml | CIP-0005 Recommendation §§1–7 | Token source refusal, byte caps, child environment, canary/cache distinction and strict/advisory handling represented; no additional blocking finding. |
| Summaries/links | CIP-0002 through CIP-0011, README, diagram index | 10/10 operative summaries present with purpose/boundary/authority/I-O/phase/landing links. 13/13 inventory entries present and linked from README/index/relevant CIPs. Existing disposition blocks retained. |
| R-CS2 F1–F3 | Shared disposition and CIP-0002 against second-review record | Identity deduplication and existing dispatcher position restored; private precondition filename removed from shared public text; rule-derived rows explicitly pending matching module vectors. |
| Outcome accuracy | TASK-261011-23nfdx_figures.md | Figure/file/spec/audit mapping and counts agree. Its broad correspondence/recovery claims need qualification by A-F1–A-F4; existence is not full semantic acceptance. |

## Evidence and limits

Reviewed the exact base-to-candidate delta (26 documentation/diagram paths) and defining local prose; re-read the finding-bearing figure content from the candidate Git tree. No code/schema changes, repository edits, commits, LOGBOOK writes, renders, tests, builds or standalone lint/validation were performed by this reviewer. No added personal workstation/session links or first names were found in the reviewed delta; pre-existing document author metadata was not rewritten.

The producer handoff validation log reports exit 0, 73 schemas, 1294 vector files, 689 tests and a Go compile-only step, with exact-command coverage 1/1 and test-case coverage unknown. This is attached prior evidence read by this reviewer, not rerun evidence and not proof of PlantUML rendering or semantic correspondence. No renderer result is claimed. Remote-main freshness was not established; this record addresses the assigned immutable CR revision.

The specific reviewer-A brief overrides generic lifecycle instructions: no set_status, accept_cr, reject_cr, handoff or commit_ack was issued. Reviewer B owns the decision. The same report is attached under the requested p1-R-CS3-review-A.md name and a task-scoped alias for discoverability.
