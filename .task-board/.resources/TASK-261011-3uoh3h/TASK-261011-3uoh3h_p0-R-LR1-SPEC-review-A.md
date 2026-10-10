# TASK-261011-3uoh3h — R-LR1-SPEC-contract-fix: reviewer A

Record-only review of CR-TASK-261011-3uoh3h-1 revision 1. Recommendation: accept the documentation change; no blocking findings. This is one of two reviews, not Change Request acceptance or integration authorization. Per p0-R-LR1-SPEC-review-A-brief.md, no accept_cr/reject_cr or final verdict status mutation was performed.

Reviewed exact base 68be5ffe2bc9ec0fc9e8a1bb8777368c55069ded against candidate tree 70083a7e9155400c9f6d8ec6ecc55299fd9a65da. Authoritative sources: curator contract table at 177cb554, audit at 48789217 (D1 R-LR1 and its A-rows), and supplied owner-decisions-20261010-audit.md. Retrieved and checked TASK-261011-3uoh3h_alignment.md through board resource get. spawn goal reports this run is not goal-bound.

## Findings

```yaml
findings:
  - severity: note
    severity_reason: "P3 informational: activation dependencies are explicitly open and permitted by the task acceptance criteria; no corrective change required."
    location: "SPEC.md sections 10.1, 10.4, 10.5; README Target one-shot exec mode"
    finding: "Intake schema/digest preimage and approved argv, completion receipt transport and ownership across exec, numeric exits, helper signatures and launch-ticket activation remain unpublished. The candidate supplies explicit options and does not claim implementation or qualification."
    disposition: "Retain these gates for subsequent implementation/publication work. They do not block this docs-only repair."
```

## Swept surfaces and alignment

All four audit rows named by R-LR1 are covered (4/4); R-LR1 does not enumerate ten A-rows. To meet the requested minimum of ten row comparisons, the following also checks distinct canonical contract-table rows. Ratios concern document review, never runtime coverage.

| Row | Changed evidence | Review result |
|---|---|---|
| A09.1 | SPEC introduction, sections 1/1.1/10.2; README front doors | Caller composes; agent curator-run verifies/executes; privileged launcher remains distinct. Interactive execution retained. |
| A09.6 | SPEC 10.1–10.5; README target mode | FD3 intake, credential binding, stop, receipt and refusal boundaries specified; unpublished choices explicit. |
| A09.10 | SPEC 1.1, 3, 10 status and 10.1–10.2 | D-EXECUTOR selected narrow mode; version publication still required; local headless v0 only; no implementation claim. |
| A10.9 | SPEC 10.3; README consumer paragraphs | D-EXECUTOR chooses curator-run first; board/host ask dispatcher and embed no broker client. |
| N6/P1 | SPEC 1.1/10 | Correct binary/repository role; current interactive contract separated from target. |
| L1 | SPEC 10.1 bound table | Helper 16 KiB and launch 80 KiB, closed JSON within 5 seconds. |
| L2 | SPEC 10.1 bound table | Decoded 48 KiB, base64 64 KiB, dispatcher 96 KiB/5 seconds and 128 KiB frame, broker 64 KiB; smallest path bound before provisioning. |
| P2 | SPEC 10.1–10.2 | Root validates generation/home/executor, resets descriptors/environment, writes receipt before dropping groups/GID/UID; fresh broker connection afterward. |
| I1 | SPEC 10.2 | Root correlation separate from per-operation effect IDs; retries reuse their own effect. |
| I2 | SPEC 10.2–10.3 | Allocation, execution claim, dispatcher and board handles distinct; plan-digest launch_id is correlation only. |
| I3 | SPEC 10.3 | Active uid/generation, never-reused generation and random 128-bit never-reused binding ID. |
| I4 | SPEC 10.5 | Target ticket fields and once-only redemption retained; signature/host-role activation explicitly open with options. |
| I5 | SPEC 10.3 | Default protected ledger locator and 0644 mode; active tuple required at bind/lease/renew; read failure refuses. |
| P5 | SPEC 10.3–10.4 | Executor/runner renews, broker refreshes; 8-second default, fail turn/no fallback; across-exec ownership explicitly requires publication. |
| P6 | SPEC 10.4 | Dispatcher/helper exec.stop, verified process identity, TERM then KILL, confirmed absence including descendants; unknown evidence cannot release or replace. |
| C1/C2 | SPEC 10.3 lane table | v1 absent source means per-home; v2 explicit source; native home distinct, read failure not absence. |
| C3 | SPEC 10.3 | Claude enrolled setup-token env channel; no refresh/file/native fallback; expiry requires restart/re-enrollment. |
| C4 | SPEC 10.3 | Later Codex external-token per-session app-server, sole auth owner, no auth.json or broker codex exec; adoption/evidence versus deferral explicit. |
| C5/C6/C7 | SPEC 10.3 | Enterprise later, API-key explicit qualified channels, Muse subscription deferred/open with options. |
| C8 | SPEC 10.3 | Metadata-only extension, protected destination binding including harness digest/policy, qualified injection, source conflict refusal, readable-token boundary, first file store 0600/0700. |
| C9 | SPEC 10.3 | Owner decision overrides old consumer ambiguity; first Claude proof, Codex/network later. |
| W2/W3 | SPEC 10.5 | D-WIRE-NAMES broker domains applied; transport distinguished from signed domain; helper concrete signature grammar not invented. |
| X1/X2 | SPEC 10.2 | D-PATH exact three-part order under plan digest; repository .agents/bin forbidden; interactive append and reserved metadata remain separate. |
| G1/G2 | SPEC 10.2 | Caller/board authorship, no agent capability; rationale/deferred SH2+ choice; no board/goals/toolchain dependency. |
| R1/R2 | SPEC 10.4 | Dispatcher revokes only registered one-shot identities after retirement; standing identities operator-only signed record. |
| D5 | SPEC 10.5 | D-DONOR-PROVISIONING selects separate component/privileges/removal; variant B stays disabled. |
| P8 | SPEC 10.5 | D-RESUME same-major/0.x same-minor, resume-failed and caller decision, no automatic fallback/seed; v0 resume still unsupported. |

The supplementary comparisons cover 31 distinct contract-table rows: N6, P1, L1, L2, P2, I1–I5, P5, P6, C1–C9, W2, W3, X1, X2, G1, G2, R1, R2, D5, P8. Coverage is 31/31 listed canonical rows, plus 4/4 R-LR1 audit rows. Owner decisions are treated as overrides only for their recorded choices; no recommendation was found promoted to a decision.

Alignment artifact: its four primary row mappings and supplementary dispositions agree with the changed sections. Its current/target separation and explicit publication questions are accurate.

## Diagram and hygiene checks

Audit R-LR1 names no launcher diagram files; audit B1 reports none in its launcher inventory. Candidate tree has no .puml/.mmd/.svg files; SPEC/README contain no embedded diagram markers or image references. Thus no existing launcher diagram needs correction. External launch figures have other repair owners; new M-LR1/M-LR2 remain P1 R-LR2.

Exact candidate delta changes only README.md and SPEC.md (435 insertions, 11 deletions); no code or tests changed. Interactive changes are scope qualifications, stale import wording correction and target pointers; existing CLI/exit behavior is retained. Reviewed added text contains no personal first names, machine-local paths, session links or board-resource links. The /opt/swarma ledger path is the canonical platform locator, not a local workstation path.

## Validation and limits

Personally reran git diff --check on the exact base/candidate pair: exit 0. Personally inspected the exact diff, candidate source inventory, target sections, README additions, pinned authority rows and alignment artifact. No tests, builds, vet, race or renderer commands ran, as the p0 brief prohibits them. The configured automatic handoff suite is sanctioned; R223 limits its host Go step to compile-only. This review does not claim a handoff suite result or runtime/negative-test qualification and does not use compile-only evidence as behavioral proof.

No new regression, decision or blocker was discovered requiring a separate logbook entry. Repository files were not modified by this reviewer. Record-only review A ends with this evidence; the review coordinator retains the subsequent review/acceptance lifecycle.
