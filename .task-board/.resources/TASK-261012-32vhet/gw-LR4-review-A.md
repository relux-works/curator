# TASK-261012-32vhet — apply-cycle2-findings-lr — reviewer A
Record-only review of CR-TASK-261012-32vhet-1 revision 1, 2026-10-12.

Base: 2297327247b7fc1fd793de52aea9fce856322204
Candidate tree: fa2083e5c7c651b6df6c66b81c6cd18d29e465d8

## Findings
```yaml
findings:
  - id: LR4-A-N1
    severity: note
    severity_reason: "P3 informational publication dependency; no launcher rework requested."
    location: "SPEC.md §§10.2, 10.4–10.5"
    finding: "Neighbouring main checkouts still contain cycle-1 contracts. Wave-4 working drafts contain the referenced account-wide stop, fate table, session permit and time rules. These draft reads do not establish publication or qualification."
    recommendation: "Retain the existing joint publication and qualification gates; reviewer B should treat upstream landing as a separate dependency, not runtime evidence."
actionable_findings: []
```

No bypass, regression or robustness defect found in this bounded documentation review. This is review evidence, not an acceptance mutation; reviewer B decides.

## Swept surfaces
| Surface | Evidence and result |
|---|---|
| F-16 | SPEC §10.2 and §10.5: standalone cross-account repository; private .git/common directory/refs/index/config/hooks; no inherited backlink, executable config or credential helpers; copy/reflink preserves privacy; immutable read-only object sharing; same-account worktrees remain. Transfer ownership and object provenance/access/retention stay open. |
| F-13 | SPEC §10.4 references user-manager §§4.4/4.7 and launch figure. All-UID signalling and quiet absence share one scope; latest-receipt fence, non-signalling completed replay, escaped descendants and unknown-evidence closure are explicit. |
| F-18 | SPEC §10.4 references user-manager §7 beside R-FD3, including process fates and replaced-daemon writer fencing. No live supervisor adoption or inferred setsid survival. |
| F-20 | SPEC §10.2 replaces four-value permit identity with session scope and preserves each physical start's claim/fence/receipt. Successor consent, previous absence, idempotency and host_stop restrictions match R-SESSION. §10.5 keeps member/transport formats unpublished. |
| F-21 | SPEC §10.4 carries authenticated/operator-controlled time source, durable watermarks, max/min comparisons, no untrusted advance, operator reset and elapsed-lifetime destruction prerequisite. Detailed implementation remains component-owned. |
| Earlier decisions | §10.4 preserves ordinary home deletion, restricted metadata export and explicit operator retention (F-1); §10.2 preserves dispatcher-only starts (F-4) and Unix-socket service from the first slice (F-5). |
| Shared names and wire scope | R-SESSION/R-FATE/R-TIME and existing R-START/R-FD3 retain supplied names and meanings. Other listed resolutions are outside this launcher brief. No newly invented wire member, refusal code or value found in the delta. |
| Figures | All 4/4 changed figures inspected: artifact-layout, components, launch-paths, launch-states. Text matches the changed contracts; diagrams/README.md links all four. Source review only, rendering unknown. |
| Revision and gap report | Dated revision at SPEC top follows existing style. TASK-261012-32vhet_gaps.md covers all 5/5 touched findings and retained publication questions accurately. |
| Public hygiene | Inspected all 7/7 changed paths and searched spec/README/figures for private paths, board-resource/session links and stale permit/worktree wording. No prohibited identity/location data found in the delta; README removes the machine-specific runner wording. This is bounded source inspection, not proof about every historical repository file. |

## Sources and verification limits
Read the three producer preconditions, reviewer-A brief, exact base-to-candidate diff, current §10 context, all changed figure deltas/index, gap report and attached revision-1 validation log. Working files matched the candidate tree when compared with git diff.

Compared neighbouring main specs and the available wave-4 Story working drafts: user manager §§4.4/4.7/7/11, dispatcher §§3.4/12, session-host §§3.2–3.5/7, and broker time/fate references. Drafts may change independently; no upstream publication status is inferred from their presence.

No tests, builds, renderers or validation gates were run by this reviewer. Accepted as existing evidence only: the attached CR validation log reports remote CI run 38283654517 successful, Ubuntu/macOS test and race jobs successful, optional self-hosted job skipped; command shard coverage 1/1, test-case coverage unknown. No independent hosted attestation or platform qualification is claimed.

The run goal query reported no active goal. No repository code or documentation was modified. LOGBOOK.md was not written: the task brief prohibits it; this attached report preserves review findings. Per the specific record-only hard rule, no status transition, accept_cr, reject_cr, handoff or commit acknowledgement is issued.
