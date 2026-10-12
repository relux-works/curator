# TASK-261012-2gkqm3 - review verdict, revision 1 (reviewer B, deciding)

Verdict: CHANGES REQUESTED (route: to-dev). One P1, a small fix. Everything else in the candidate passes.

Base eddad09, candidate tree e1344bd, 12 docs/diagram paths. My checks 1-7 were done before I read the reviewer A record (gw-CS3-review-A.md).

## Defect (P1)

D1. `cips/diagrams/join-handshake.puml` line 72 (the note in the join-handshake figure) still says "Replay: ... frames meet connection-scoped ids and op_key caches." This is the claim F-7 retires. CIP-0009 D6 now says connection-scoped IDs and the ten-minute output cache neither prove an effect absent nor authorize a retry. The figure lists exactly those as the replay defence.
- The figure is linked from CIP-0009 D4/D6 and from the README figure index (Join handshake -> D4, D6, D9).
- The brief requires: "Update every figure that shows a changed behaviour, so no contradicting copy remains." The AC requires: "figures match the text".
- The gaps file lists eight figure sources and says no contradicting copy remains. That is untrue for this file.
- Required fix: change the replay line to match D6 and the remote-worker-turn figure. Example: "Replay: invited cert meets its tombstone; collection proof/retry format pending; later frame effects rely on durable binding-generation identity/digest claims with authenticated admission floors (CIP-0009 D5), not on connection IDs or the output cache." Add a one-line revision comment in the file's existing style. Add join-handshake.puml to the gaps file's F-7 figure list. Do not invent fields or codes.

## What passes (checked myself)

1. Brief items: F-7 (CIP-0009 D5/D6, Security, Test plan, open question 12, remote-worker.md, remote-worker-turn.puml); CIP-0011 for F-5/F-3/F-8/F-11/F-1; F-4 ticket text removed. All applied.
2. Owner decisions F-1/F-4/F-5: retirement deletes the home, with a closed-schema metadata export only and a separate explicit retention that disclaims confidentiality. Ticket, host role and redemption are gone. Socket service from slice 0, no command mode, sudoers, SUDO_UID or conversion. A grep of cips/, protocol/remote-worker.md and README for ticket, SUDO_UID, command mode, sweep timer, redeem and archive finds only unrelated hits. Dispatcher v0/v1 phrasing is gone.
3. Shared names are used unchanged: effect.observe with six states, revocation.check with not_revoked/revoked/unavailable and context prepare/dequeue/start, host_start with the four-value tuple and the host admission permit, R-DELEG (requester-signed, registration is not activation), R-MAINT. F-6 correctly untouched.
4. No invented wire fields, codes or durations. Open questions say "would need ..." (CIP-0009 q12, CIP-0011 q4/q5, cross-CIP rows).
5. Security properties survive: unknown stays pending/unknown_result or unavailable and is never absence; no automatic re-execution after cache expiry, rollback or not_recorded; descendants carry the requester in their chain; the verdict is point-in-time and the final bind/lease/renewal checks stay; a fenced restore returns unavailable.
6. Figures: remote-worker-turn, retire, lifecycle, donor-side (+dsl), layout, variants and credential-source-lifecycle are consistent with the text. alt/loop/break/end blocks are balanced and participants exist. Only join-handshake is stale (D1).
7. Hygiene: added lines contain no first names, hosts, local paths, session links or board links.
8. Gaps file: complete and truthful apart from the D1 figure omission. CR validation log is green (exit 0, 689 tests OK).

## Notes (P2/P3, not blocking)

- N1 (P3): `platform-contract-disposition.md` row "User-manager helper / launcher requests" still says "Configured sudo caller/role policy in v0/v1". This is the pre-existing helper policy (dispatcher-to-root), not the removed host role. A later editor could make it explicit that no host role exists.
- N2 (P3): reviewer A's note stands: neighbouring module specs still describe the old behaviour until the parallel module work lands.

## Reconciliation with reviewer A

A found no actionable defect and no P0/P1. A's figure sweep covered the eight changed files but not unchanged figures that restate the retired replay claim, so it missed D1. I checked it with a repository-wide grep for op_key, unknown_result and cached result.
