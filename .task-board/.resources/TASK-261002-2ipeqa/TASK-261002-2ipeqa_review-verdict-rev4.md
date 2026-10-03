# TASK-261002-2ipeqa — accept-rc14-candidate-suite: revision 4 review

Verdict: accepted. No open findings. Revision 4 candidate tree 9706982aa19b64fa55b95123e68a34b3640818ca equals the revision 3 reviewed tree exactly. Base: 64345d71aa539ef71b9f0b4d3fb9b28d1c9d3fb7.

Previous finding R3-F1 (scope mismatch) is resolved by the binding rc14-lockstep-review4-note.md and rc14-lockstep-noop.md expressly accepting all five files in the rev2-to-rev3 delta. This is scope reconciliation, with no code change. The prior review found no product defect. Repeat-of: R3-F1 resolved; no recurring mechanism or new finding.

| Swept surface | Result and evidence |
| --- | --- |
| Revision identity | Full supplied rev4 OID equals rev3 OID in TASK-261002-2ipeqa_review-verdict-rev3.md. git rev-parse 9706982a^{tree} and e4baf6fa^{tree} both return that OID; git diff --exit-code 9706982a e4baf6fa exits 0. |
| Scope | Reconciled five-file delta: gap, snapshot RunOutcomes adapter, regressions, 12 lifecycle pins (8 historical), changelog. Full base-to-candidate remains 17 paths; no new delta. |
| Candidate evidence | Fresh gh run view 37017428885 --json headSha,conclusion,jobs succeeds (exit 0). Head e4baf6fa3fe608e4878c8a733939b1293edf0b27, conclusion success. Candidate suite jobs AND Candidate suite + platform-case gate steps succeed on Ubuntu, macOS, Windows: 3/3. |
| Counts and digest | Reuse exact-tree rev3 independent immutable-corpus recount: digest 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5; marker 29, fragment 36, lifecycle 6/3/4/4. Prior reviewer measured 6/107 families, producer audit covers 107/107; ledger total 1887. No new full recount claimed. |
| Gaps and fixed cases | Reuse rev3 sweep: 10 rc.14 gaps, only new gap byte-exact-snapshot owned by TASK-261002-1foyf3 with cut-over reason; seven previously fixed cases absent. |
| Snapshot and negatives | Reuse rev3 review of production extraction and byte checks, outcome adapter, stale-gap/unlisted-failure regressions, and producer narrowing mutants. No new mutation run claimed. |
| Older suites | rc.13 unchanged; prior candidate dispatch retained, with eight historical lifecycle pins explicitly authorized by reconciled scope. Full historical production replay remains outside the evidence. |
| Spec exact command | Reuse rev3 evidence: exact six-package Implementations argv exit 0; pinned Python checker on exact producer stream and hosted stream passes 8/8 claims each. |
| Switches and hygiene | Fresh protected-path diff against base is empty. Writer false and SPEC_PIN 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065 independently read. LOGBOOK unchanged; git diff --check exits 0. |
| Default matrix | Fresh job listing confirms default 3/3, race Linux/macOS, lint, interop and self-tests success. Only optional rose-air skipped. |

Evidence reuse: TASK-261002-2ipeqa_review-verdict-rev3.md (read in full), TASK-261002-2ipeqa_revision3-results.md and TASK-261002-2ipeqa_revision3-evidence.tar.gz as referenced by that review; https://github.com/relux-works/curator/actions/runs/37017428885. Binding review4 instructions expressly permit this exact-tree reuse.

This reviewer performed read-only Git, gh and board queries. Successful verification calls exited 0. Exploratory unsupported board projections and a missing optional reviewer reference were not treated as evidence; corrected board get projection succeeded. No Go build/test/vet/run, local lint/build or mutation tests were run in revision 4, and no syspolicyd duration/crash pair is claimed. Prior review records the earlier twice-stalled marker/scriptworker wrappers exiting 130 with underlying Go exits unknown; no retry here. Hosted validation remains the full-matrix arbiter.

Queried spawn goal: run is not goal-bound. Live checklist is fully checked. Repository code and LOGBOOK were not modified. Sole verdict branch: accept_cr revision 4, routing to integrating; no done transition or commit_ack.
