# TASK-261012-32vhet — reviewer B verdict, CR-TASK-261012-32vhet-1 revision 1

Verdict: ACCEPTED. Docs-only change request, 7 paths, base 2297327, candidate tree fa2083e.
Checks 1-7 were done and written down before reading reviewer A's report. A's report was then reconciled.

## Checks
1. Brief items. F-16 is in SPEC 10.2, with the transfer question in 10.5. F-13 and F-18 are in SPEC 10.4. F-20 (R-SESSION) and F-21 (R-TIME) were applied as the AC requires: 10.2, 10.4 and 10.5. All are reported truthfully in the gap report. Nothing was claimed as already closed.
2. Owner decisions.
   - F-16: standalone repository with private .git, common directory, refs, index, configuration and hooks. No inherited backlink, executable config or credential helpers. Copy/reflink keeps the rule. Object sharing is immutable and read-only. Linked worktrees stay within one account. The transfer owner is left open.
   - F-20: the permit is session-scoped for one logical_session_id. Session lifecycle authorisation is unpublished. Each physical start keeps its (execution_id, controller_epoch, op_id) interlock. A successor host_start is idempotent per (logical_session_id, op_id) and needs the previous claim confirmed absent. host_stop is host-principal only, under the permit, goes through the journaled E-stop, is host-requested, differs from owner cancel and confers no start authority.
   - F-21: authenticated or operator-controlled time source, durable watermarks, max for expiry, min for not_before, no untrusted advance, operator reset, and a boot-inclusive elapsed-time gate on destruction.
   - F-1, F-4 (dispatcher-only starts) and F-5 (Unix-socket service) are intact. The old four-value permit wording and the "fence withdraws permit" wording are removed, with no stale copy in SPEC, README or the figures (grep-verified).
3. Shared names. R-SESSION, R-FATE and R-TIME are used with their resolved meaning. R-FD3 and R-START are untouched. The other resolutions concern other repositories and are not misused.
4. Nothing invented. No new wire fields, refusal codes or values. The member format, transport, transfer protocol and object-sharing rules are named as unpublished and kept in the open-question list ("would need ...").
5. Security properties. Fail-closed on unknown evidence is retained ("Unknown/contradictory evidence permits neither claim release nor replacement"). No second launch after an ambiguous call (same effect or receipt). Successors need explicit owner authorisation, and no widening is stated. Clock steps can only narrow authority. Old receipts cannot stop successors. Persistence denial and the idle-zero rule are left as upstream qualification obligations.
6. Figures. artifact-layout, components, launch-paths and launch-states plus the figure index all match the text, and all four are linked from diagrams/README.md. PlantUML syntax is conservative. Rendering is unknown (no renderer on this host).
7. Gap report. It is complete and accurate: 5/5 touched findings, an open register, and honest disclosure of the README CI runner wording removal and the board-prerequisite anomalies. CR validation log: remote gate success on ubuntu/macOS test and race; the self-hosted job was skipped; test-case coverage unknown. That is CI evidence, not documentation qualification.
8. Hygiene. The new text has no first names, host names, absolute paths, session links or board-resource links. The delta also removes a host-name-bearing CI runner label from the README. Pre-existing task IDs in older revision-note rows are not touched by this CR.

## Reconciliation with reviewer A
A found 0 actionable findings and 1 note (upstream publication dependency). I agree. No P0/P1 anywhere.

## Notes (P2, tb-R226; no rework)
- N1. The launcher repeats the same F-16/F-13/R-TIME summary in the SPEC, the README and the notes of all four figures. That is a maintenance-drift risk, and a future revision could let the figures cite the SPEC sections instead.
- N2. F-20 permit withdrawal, unknown-permit-state-stays-closed and per-use revocation.check rows are dispatcher-owned. The launcher text does not restate them and only says "without widening". A cross-reference line in 10.2 would make the fail-closed property visible here.
- N3. The spec refers to user-manager `diagrams/launch.puml` by relative name, next to this repo's own launch-paths.puml. That is slightly ambiguous but already attributed to the user manager.
- N4. Reviewer slip: an accidental `git checkout --detach` in the worktree was reverted immediately with git switch. Branch, HEAD OID 2297327 and the 7 uncommitted modifications were verified intact. No content was changed.
