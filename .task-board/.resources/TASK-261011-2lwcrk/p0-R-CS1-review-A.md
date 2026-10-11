# TASK-261011-2lwcrk — R-CS1 contract fix — reviewer A

Reviewed Change Request CR-TASK-261011-2lwcrk-2 revision 2, base a60fd34173abc82401b5aee00c6204dce2456997, candidate ce2b96e7be1d323fad01a747fa4da37a078a08da.

Assessment: no blocking findings; recommend acceptance to the deciding reviewer. This is RECORD-ONLY reviewer A evidence, not a persisted acceptance verdict. The task-specific p0-R-CS1-review-A-brief.md explicitly prohibits accept_cr/reject_cr. Its rev1 wording is stale relative to the supplied rev2 candidate; this review covers rev2 only. No repository modifications or lifecycle verdict mutation were made. The run is not goal-bound (spawn goal queried).

## Findings

```yaml
findings:
  - id: A-N1
    severity: note
    severity_reason: "P3: evidence limitation, not a delivery blocker for this docs-only scope"
    surface: "protocol publication and five diagram sources"
    finding: "Closed schemas and channel/backend ownership remain explicitly pending. Diagram sources were manually reviewed but not rendered or syntax-tool validated. Neither publication nor the green repository suite establishes remote-worker runtime qualification."
    requested_action: "Retain these publication and qualification limits in the deciding review."
```

## Swept surfaces

Sources read: curator contract table at 177cb554, audit at 48789217 (D1/R-CS1, relevant A rows and B2), and supplied owner decisions dated 2026-10-10/11. Contract D1-D5 and the owner overrides govern the assessment. Coverage: 10/10 named A rows and 5/5 named B2 source files, manual documentary coverage only.

| Audit row | Candidate evidence | Assessment |
|---|---|---|
| A04.4 | CIP-0009 D0/D4; protocol deployment boundary; deployment/C4 key descriptions | Three separate keys; supported A/C Secure Enclave P-256 signing/SSH, separate X25519 encryption, B Ed25519 files/no Keychain. No universal Ed25519 recipe. |
| A04.5 | CIP-0009 D4-D7/D12 and connection tests; protocol Trust and SSH | Host CA plus user CA; TXT host-CA publication, certificate principals, renewal/KRL, explicit session closure and overlapping CA rotation. Durations stay approximate policy examples. |
| A11.2 | CIP-0009 D0-D3/security/cut list; CIP-0008 R3-R8/security | A first/no admin/no OS confinement; B one admin request, separate identities, no Keychain; C donor-UID sandbox fallback. Separate donor provisioner, B disabled pending specification/qualification. A availability after B/C remains an explicit two-option owner question. |
| A11.5 | CIP-0009 D4/D6/decision 2/cut list/tests; protocol trust section | Local DNSSEC default with DoH transport fallback; invitation pins only for declared no-DNSSEC environment. Read/validation failure never selects fallback; coexistence requires agreement; records per environment. |
| A11.6 | CIP-0009 D4; protocol publication register | Challenge/proof/operator admission, atomic invited consumption and retries, three keys, encrypted signed package without carrier secret. Invitation/join/package schemas each pending, including open-application binding/retries. |
| A11.7 | CIP-0009 D9; protocol join limits/register | Three verbs, no signing key or command execution; existing owner policy limits retained (10/minute, 100/day/address, five challenges, one use, 72 hours, 16 KiB, five minutes, 50 pending). API/error formats pending. |
| A11.8 | CIP-0009 D10/D4; protocol deployment boundary | Human-only consent; TTY/environment/typed confirmation explicitly heuristic, not proof of identity. |
| A11.9 | CIP-0009 D5/D12; protocol register/trust | Protected forced executor, certificates/CA/KRL, no forwarding/PTY/rc/environment; bounded framing, idempotency/caps/leases/cleanup requirements preserved. Backend/channel closed formats pending with explicit repository-owner options. |
| A11.10 | CIP-0009 D3/D5/D11; protocol mailbox/lifecycle; turn source | Policy intersection, donor admission, durable cursors/dedup/application receipt, donor signs final bytes, project carrier distinct; disconnect behavior and later federation phase retained. |
| A11.11 | CIP-0009 D7/D8/tests; protocol lifecycle; retirement source | Independent state axes; A/C preserve donor account; B cleanup separate/pending. Retirement revokes certificates/closes sessions but preserves identity; separate operator revocation. Offline purge reports unknown remote state. |

| B2 source | Review result |
|---|---|
| donor-side.dsl | B view explicitly scoped/disabled; A/C notes; donor provisioner separate from user-manager; trust branches and certificate SSH; pending formats disclosed. |
| donor-first-deployment.puml | Explicit A/C/B provisioning branches and disabled B gate; variant keys, invitation-file invocation, trust refusal, workplace before package; pending schemas marked. |
| join-handshake.puml | DNSSEC/default and declared pin branch; no downgrade on failure; invited/open separation; CA package and certificate renewal/rotation; schemas pending. |
| remote-worker-turn.puml | B shown with A/C limits stated; donor-originated channel, host/user certificates, protected policy, caller-composed plan, tool/admission/signing/disconnect flows. |
| remote-worker-retire.puml | Canonical project account owners, separate donor provisioner; stop/leave/purge and suspend/retire/revoke distinct; KRL plus session closure; variant cleanup and offline unknown. |

The alignment outcome TASK-261011-2lwcrk_alignment.md maps all 15 rows truthfully at the document/section level. Its validation discussion describes the pre-handoff state; the subsequent rev2 validation log supplies the current green result. Pending schemas are not presented as already implemented operations. D-DONOR-PROVISIONING resolves contract D5 without extending user-manager. D-WIRE-NAMES is applied to enrollment objects; transport/backend repository ownership is explicitly open rather than invented. D-IDENTITY preserves donor identity on retirement. D-GOALS, D-PATH, D-EXECUTOR and D-RESUME introduce no additional change obligation in this bounded donor reconciliation; no conflicting new self-goal, managed PATH, one-shot composition or native-resume policy is introduced.

## Verification and limits

- Exact delta: eight documentation/diagram-source paths, no product code or schema implementation. All eight worktree file byte sequences were compared with the candidate Git blobs and matched, including the untracked protocol file. The Git diff reporting that untracked file as deleted when comparing to the candidate is an index/worktree accounting artifact, not a missing candidate file.
- Direct standalone git diff --check: exit 0. Direct base-to-candidate git diff --check: exit 0, including the new protocol blob.
- Accepted attached evidence only: TASK-261011-2lwcrk_change-request_rev2-validation.log reports 73 schemas, 1294 vectors, 689 passing tests, Go compile-only per R223, exact-command shards 1/1 green, exit 0. Test-case coverage is unknown. No tests/builds were rerun by this reviewer.
- Producer alignment explicitly discloses historical no-index exit 1 and previous failed suite. Neither is counted as a current pass.
- No renderer, diagram syntax validator, provider qualification, deployment or external source freshness claim. No new first names, personal machine paths, session links or board-resource links found in the candidate documents; generic specified deployment paths are contract examples.
- No commit, branch operation, source edit, accept_cr, reject_cr or done transition. Reviewer A record is complete; deciding review remains responsible for the CR verdict under the explicit two-reviewer brief.
