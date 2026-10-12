# TASK-261012-2uuu3e — apply-cycle2-findings-cs — reviewer A

Record-only review of CR-TASK-261012-2uuu3e-1 revision 1.
Base: `601cbb090851a255357fce95575e0bfacd9c37f9`.
Candidate tree: `9c1b487b43e79b66047abcf7070774fb461ca185`.

No blocking defect found in the bounded review below. This is not acceptance: reviewer B owns the decision. No status change, accept_cr, reject_cr, code edit, commit, test, build or renderer invocation was performed. The specific reviewer-A hard rule overrides the generic lifecycle instructions. The run goal query reported no active goal.

## Findings

```yaml
findings:
  - id: CS4-A-N1
    severity: note
    severity_reason: "P3 informational: publication and runtime qualification remain explicitly pending; no defect or acceptance blocker established."
    location: "CIP-0009 questions 13–14; CIP-0011 questions 4, 6–8; protocol/remote-worker.md publication register"
    evidence: "Workplace backend shape, service deployment/socket authentication, safe private-home access, CA renewal grammar, session authorisation/permit encodings, renewal/time bookkeeping and networking qualification are named as unpublished. Both CIPs remain Draft."
    recommendation: "Preserve these activation gates through integration; do not interpret this documentation review or the existing spec gate as runtime qualification."
```

## Swept surfaces

| Surface | Evidence and result |
|---|---|
| Brief item 1 / F-19 | CIP-0009 D4.7, D5, D12, security, implementation step 2, test plan and questions 13–14 place swarma-rb-exec under a service principal. Workplace UID runs untrusted jobs/runner; fixed SSH shell is relay-only. Claims, floors, caps, audit, mailbox controls and CA custody are outside workplace authority. Helper v1 activation/applied state and loopback apply to all variants, including read-only. Execution cleanup uses dispatcher/root receipt/exec.stop, with runner blast radius and unknown sibling claims explicit. Remote-worker register labels both design wires UNPUBLISHED. |
| Brief item 2 / F-15 | CIP-0011 §4, slice 0 and test plan distinguish control readiness from forward admission; authenticated observation/reconcile and sound current-execution stops remain reachable. Record-local unknown retains identities and capacity; unrelated work needs complete accounting/confinement; shared uncertainty/restore/maintenance stay global barriers. |
| Brief item 3 / F-27 | CIP-0011 §6 cites N15 as binding owner and explicitly replaces [bindings.profiles]. Caller text is not authority; dispatcher pins machine policy; fw.apply checks served policy. N15 details are not copied into a new local schema. Its under-review/not-landed designation follows the assigned brief; no remote landing-status inference was made. |
| Brief item 4 / F-14, F-21, F-24 | CIP-0011 §4 specifies E-rebind, same uid/generation binding replacement, old tombstone, fresh prepare verdict, existing chain principal, no widening/account recreation. R-TIME gives authenticated/operator-controlled source, durable broker/dispatcher watermarks, max expiry/min not_before, trusted advancement/operator reset and elapsed-time destruction gate. §6 gives every non-root service a v1 egress policy, broker proxy-only and dispatcher/host deny-all, with helper §8 owning rows. |
| F-16 owner decision | CIP-0009 D5 and CIP-0011 §4 require standalone repositories/private Git administration across accounts, fresh configuration/hooks, and no linked-worktree backlinks; same-account worktrees and board workflow remain. Transfer ownership remains open rather than invented. |
| F-20 owner decision | CIP-0011 §4, test plan, question 4 and shared disposition use session lifecycle authorisation and session-scoped host admission permit. Physical starts retain claim/fence/receipt; successor idempotency and previous absence are explicit. host_stop is configured-host-only, permit-bound, journaled E-stop/exec.stop, host-requested distinct from owner cancel, and grants no start authority. Owner ancestry, withdrawal, ordinary lifetime/budget bounds remain. |
| Earlier F-1/F-4/F-5 | Ordinary retirement still deletes home after confirmed absence/key cleanup with only closed-schema service metadata export; explicit retention retains confidentiality warning. Physical starts remain dispatcher-owned. Unix-socket dispatcher service remains slice 0. No superseded tuple-bound permit or same-UID executor alternative remains in the reviewed target text. |
| Shared vocabulary | R-SESSION, R-REBIND, R-TIME, R-WORKPLACE, R-SVC-EGRESS and R-NETBIND retain assigned meanings. R-OBS six states, R-VERDICT contexts/values, R-DELEG and R-MAINT are preserved. R-FATE is referenced without redefining its table. R-OUTCOME, R-LOCK, R-GEN and R-FD3 are not newly republished by this CIP delta; their owning module work remains separate. No new wire field, refusal code or enum value was found in the delta. |
| Figures, 10/10 changed sources | Read credential-source-lifecycle, donor-first-deployment, donor-layout, donor-lifecycle, donor-side.puml, donor-side.dsl, donor-variants, join-handshake, remote-worker-retire and remote-worker-turn. Ownership, v1 gate, start/stop and open backend scope align with text. All ten are linked in diagrams/README.md. Source review only; rendering unknown. |
| Revision notes | CIP-0009 revision 6, CIP-0011 revision 5, remote-worker revision 3 and shared disposition revision 3 explain changes at top. Changed diagram sources/README carry revision notes. CIPs remain Draft. |
| Gap artifact | TASK-261012-2uuu3e_gaps.md accounts for all eight touched findings (F-19/15/27/14/21/24/16/20), sections, figures and open work. It does not falsely claim any finding already closed or any runtime/renderer qualification. Its handoff prediction is explicitly deferred to the attached validation log. |
| Public hygiene, 15/15 paths | Read the exact candidate delta and inspected changed sources for personal/concrete-host/session/board-resource/local-workstation disclosures. No such disclosure found. Selected-marker scan over all 15 candidate files produced one false positive in existing prose 'integrity/home/profile/lock/auth/grant'; no actual private-path hit. Platform installation roots and generic host placeholders are contract examples. This is bounded manual/marker inspection, not an exhaustive personal-name detector. |

## Neighbouring contract checks

Read the relevant cycle-2 source passages in the available sibling worktrees: dispatcher §§3.4, 5, 8, 12 (R-SESSION, R-REBIND, R-NETBIND, R-TIME/workplace); user manager §§4.3, 7, 8 (workplace, launcher/stop, fate table, applied state/service rows); broker §§7.2, 8, 15 (replacement, lease/time rules), alongside their baseline specifications. Read network-profiles N15 directly, including root-owned profile-name/lock-digest binding, served-policy relation and service egress. These support the CIP alignment; sibling working text is not evidence of landed publication. Baseline main copies still contain earlier contracts, which the shared disposition explicitly treats as superseded by owner decisions. No implementation availability is inferred.

## Verification and limits

- Exact base-to-candidate diff covers 15 files, 215 insertions and 100 deletions. Working-tree-to-candidate tracked diff was empty when checked, so source reads matched the candidate.
- Accepted already-attached evidence only: TASK-261012-2uuu3e_change-request_rev1-validation.log reports 73 schemas, 1294 vector files, 689 tests OK, Go compile-only, exit 0; exact-command-shard coverage required=1/green=1/failed=0/missing=0. Test-case coverage is explicitly unknown. I did not rerun this gate and do not infer security/runtime coverage from it.
- No tests/builds/renderers were run, per the brief. New negative test-plan rows are requirements, not executed results. No host account/firewall/provider qualification was performed.
- No LOGBOOK.md was written. Review evidence is attached on the task as requested.
- Reviewer B must make the lifecycle decision; this report intentionally leaves board status and CR state unchanged.
