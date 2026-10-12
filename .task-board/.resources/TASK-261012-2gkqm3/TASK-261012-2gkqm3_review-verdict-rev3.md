# TASK-261012-2gkqm3 - review verdict, revision 3 (reviewer, deciding)

Verdict: ACCEPTED. No P0/P1 remains.

Base eddad09, candidate tree ca23e32 (the working tree equals it). 13 docs/diagram paths. `git diff --check` over the delta: exit 0. No tests or builds were run, per the brief.

## Rev1 defect D1 (P1): fixed
- `cips/diagrams/join-handshake.puml` line 73: the replay note no longer lists connection-scoped ids and op_key caches. It now says later frame effects rely on durable binding-generation-scoped operation-identity/canonical-argument-digest claims with authenticated admission floors (CIP-0009 D5), and that connection IDs and the output cache do not authorize retry mutation. The tombstone and pending collection-proof/retry wording is kept. No field or code is invented.
- A revision comment was added at line 2, in the file's existing comment style.
- Repository-wide sweep for op_key, unknown_result, cached result, connection-scoped and output cache: the remaining hits in remote-worker-turn, donor-side.dsl/puml, donor-layout and donor-variants all separate the durable claim from the output cache and forbid automatic re-execution. They match CIP-0009 D5/D6.
- `TASK-261012-2gkqm3_gaps.md` now lists join-handshake.puml in the F-7 figures. It states that the rev1 "no contradiction" claim was unsupported, and it describes the bounded sweep honestly.

## Checks 1-7, rechecked on the candidate
1. All brief items are applied: F-7 (CIP-0009 D5/D6, Security, Test plan, open question, remote-worker.md, remote-worker-turn.puml and the other figures); CIP-0011 for F-5/F-3/F-8/F-11/F-1; F-4 ticket text removed in the CIPs and the cross-CIP table.
2. F-1/F-4/F-5 are applied exactly. A repository grep for ticket, SUDO_UID, command mode, sweep timer and redeem finds no hits in the CIP scope. Archive hits are unrelated protocol text. Retirement deletes the home and exports only closed-schema metadata. Retention is a separate explicit operator operation that disclaims confidentiality.
3. effect.observe with its six states, revocation.check, host_start with the four-value tuple and host admission permit, R-DELEG and R-MAINT are used unchanged. F-6 is untouched.
4. Unpublished formats are marked "would need ..." in the open questions and the cross-CIP rows. No new wire fields, codes or durations.
5. The fail-closed properties survive: unknown stays pending, unknown_result or unavailable; no automatic re-execution; descendants carry the requester in their caller chain; the verdict is point-in-time and the final bind/lease/renewal checks stay; a fenced restore returns unavailable.
6. Figures are consistent with the text. The changed figure list matches the nine changed figure sources.
7. Hygiene: the added lines contain no first names, hosts, local paths, session links or board links.

## Notes (P3, not blocking; tb-R226)
- N1: the `platform-contract-disposition.md` row "User-manager helper / launcher requests" still says "Configured sudo caller/role policy in v0/v1". This is the pre-existing dispatcher-to-root helper policy, not the removed host role. A later edit could make that explicit.
- N2: neighbouring module specs still describe the old behaviour until the parallel module work lands.
