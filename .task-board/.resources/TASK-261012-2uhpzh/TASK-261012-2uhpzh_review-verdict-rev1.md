# TASK-261012-2uhpzh — reviewer B verdict, CR rev 1: ACCEPTED

Base 9ee6e38, candidate tree 8b07b30 (working tree identical to the candidate). Docs only, seven files. My checks 1–7 were done before reading reviewer A's record. No tests or renderer were run (none available or allowed); CR validation is accepted from the handoff log only.

## Checks
1. **Brief items.** Item 1 (F-6/R-FD3): SPEC §10.1 references the user manager's single descriptor lifecycle and restates no order. Executor gets only caller-owned 0–2 plus read-only FD3, with no service, broker or root handle. Staging requirement is stated (completely staged bounded immutable bytes preferred; a pipe only with published capacity/feeder/deadline/cleanup; never wait for a reader that cannot start). 48/80 KiB bounds kept. Intake/digest/framing/read-deadline gates kept, and joint FD3-mechanism publication added as an activation gate plus a new open question. Item 2 (F-4): `ticket`, redemption, helper host role and SH2 ticket schema are gone from SPEC, README and all four figures (grep clean). Item 3 (F-5): the stdin command-mode request row, the "v0 detached supervisor / v1 daemon later" text and the 96 KiB command bound are removed; the dispatcher is a Unix-socket service from the first slice, with no command fallback.
2. **Owner decisions.** F-1: home deleted after quiesce and declared-key cleanup; closed-schema service-generated metadata only; retention is a separate operator operation with no confidentiality guarantee; the positive-classification wording is removed ("No general home-secret classifier or default home archive is authorised"). F-4 and F-5 as above. No superseded alternative left.
3. **Names.** R-FD3 and R-START used unchanged (`host_start`, `dispatch-op/1`, the four-value permit binding, cancel-before-call, same-effect resolution, never launch twice, fresh authenticated channel, epoch registration and readiness before delivery, native session-lock check kept). R-OBS, R-VERDICT, R-DELEG and R-MAINT fall outside the launcher and are not redefined. No new wire field, refusal code or value; the 128 KiB frame is the previous v1 value renamed. Unpublished formats are named as open questions (FD3 mechanism, export schema, host_start transport/permit/fence/crash ordering, helper v2 signatures).
4. **Security properties.** Fail-closed on unknown survives (unknown process evidence is not success or permission to release; host failure retains the claim and closed fence until reconciliation; ambiguous calls resolve the same effect and never relaunch). Cancel before the call prevents it. Launch never becomes a grant: only the dispatcher invokes root, and the host has no launch authority. Revocation (R-VERDICT) and ancestor revocation (R-DELEG) are dispatcher/broker contracts and are untouched here, so nothing is weakened.
5. **Figures.** All four sources plus the figure index updated; text and figures agree. PlantUML alt/opt/end and note/end note balanced in launch-paths (18/18, 4/4); participants `targethost` and `targetrunner` declared. Rendering unverified (no renderer).
6. **Gap report.** Accurate against the diff; honest about unrun tests and unrendered figures.
7. **Hygiene of the change.** No added first names, local paths, session or board links. See note N-1 for inherited text.

## Reconciliation with reviewer A
- LR3-A-1 (CI lane host name in README 536–538 and CHANGELOG 157): inherited, not in this delta, and it names real workflow identifiers (job name and repository variable), so a prose-only change would misdescribe the config while a config rename is out of scope here. Not a P0/P1; recorded as follow-up N-1.
- LR3-A-2 (neighbour specs still show the old text): expected for parallel waves; activation stays gated. Not a defect of this CR; carry to integration review.

## Notes (P2, tb-R226, none blocking)
- N-1: follow-up task to rename the optional CI lane and its variable together and scrub prose/CHANGELOG.
- N-2: launch-paths.puml shows the host admission permit being created upon receipt of `host_start`, while SPEC §10.2 / R-START say `host_start` refers to a permit created when the start was admitted. Align the figure wording next time the figure is touched.
- N-3: README still says "Caller/sudo exit is not a stop receipt"; "sudo" now means only the dispatcher-to-launcher entry. Harmless, can be tightened later.
- N-4: README paragraph has cosmetic line wraps from the edit.

P0/P1 remaining: none.
