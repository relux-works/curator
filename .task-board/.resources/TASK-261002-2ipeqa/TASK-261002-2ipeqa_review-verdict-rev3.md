# TASK-261002-2ipeqa — accept-rc14-candidate-suite: revision 3 review

Verdict: changes_requested. Route to-dev for scope reconciliation; no demonstrated product defect and no external blocker. No repository source modified.

Reviewed base 64345d71aa539ef71b9f0b4d3fb9b28d1c9d3fb7, previous tree 7d4914a0ad91333b90923db62547b348c50d6389, revision-3 tree 9706982aa19b64fa55b95123e68a34b3640818ca.

## Finding R3-F1 (P2): binding delta restriction does not match the candidate

Location: .github/ci/conformance-case-counts.tsv:148 and :251; internal/interop/environments/snapshot_acquisition_test.go:77; internal/conformancecoverage/rc14_test.go:47.
Repeat-of: none. Mechanism: review scope mismatch, distinct from revision-2 missing candidate evidence.

The latest binding rc14-lockstep-review3-note.md requires the delta to be ONLY the one rc.14 gap row plus exact count adjustment, with nothing else changed. Actual rev2-to-rev3 diff spans five files: 12 new lifecycle count rows (four for each of three candidate digests), one rc.14 gap, snapshot RunOutcomes adapter, expanded regression tests, and a changelog amendment. In particular, eight rows alter the older candidate tables. Thus the requested scope assertion and older-suite-unchanged assertion cannot truthfully be made.

The producer disclosed this expansion and supplied a credible explanation: actual hosted rev2 streams exposed missing lifecycle pins, and the old snapshot test fatals cannot be converted to known-gap accounting by a TSV row alone. Independent recount supports the new rc.14 lifecycle counts 6/3/4/4; the adapter preserves hard failures for extraction and byte checks. I found no product defect in these additions. Nonetheless the latest explicit review restriction remains narrower than the implementation.

Required next action: reconcile the binding review brief with the actual five-file delta (recommended: expressly include the necessary adapter, regressions and lifecycle pin repairs, including historical candidates), or split the out-of-scope changes through the normal producer/reviewer flow. Do not blindly remove necessary pins to satisfy the wording. An updated scope can reuse this exact-tree green evidence; no additional local stalled-test retry is requested. This is ordinary review rework, not a human-only blocked decision.

## Swept surfaces

| Surface | Result / bound |
| --- | --- |
| Previous P1: candidate lane absent | Fixed. GitHub run 37017428885 succeeds; gh run view jobs confirms 3/3 Candidate suite jobs and their test steps actually completed successfully on Ubuntu/macOS/Windows. Optional rose-air skipped. |
| Tree and suite attribution | Run head e4baf6fa3fe608e4878c8a733939b1293edf0b27 resolves locally to exact tree 9706982aa19b64fa55b95123e68a34b3640818ca. Downloaded Ubuntu log records spec e3a88cedba7a844c594348ee89654471db086f98, digest 6f832d813efc768ea154a7d5076b512ab4be6aa9409d92e11469d21ea9bc69f5, candidate root and full-root enforcement. Go overall exit=0 and platform gate exit=0. |
| Corpus counts | Python read of immutable spec Git objects independently hashes manifest to expected digest; recounts marker-v3 schema 29, fragment-v3 schema 36, lifecycle arrays 6/3/4/4. Independently recomputed 6/107 families; summed ledger 107 families / 1887 entries. Producer 107/107 audit is accepted supporting evidence, not my full recount. |
| Gaps and fixed cases | Exactly one new rc.14 row, byte-exact-snapshot, exact cut-over owner/reason. Ten rc.14 gaps total. Remaining nine identical to rev2; seven previously fixed cases remain absent. Published snapshot count remains one, correctly treating autocrlf variants as subchecks. |
| Snapshot gate / negatives | gitops.Extract and byte/path/line-ending checks remain driven. Only final hash mismatch returns a known-gap observation; failed child tests remain failed. Reviewed exact case/suite, stale passing-gap and unlisted failure tests. Producer archive records two narrowing mutants exit 1, restored checks exit 0; not rerun by reviewer. |
| Older suites | Dispatch unchanged from rev2; rc.13 rows unchanged. Existing historical candidate rows retained but four new rows each, so byte-for-byte unchanged claim is false. Producer historical policy runs exit 0; full historical production replay not performed. |
| Spec exact check | Read e3a88ced implementations.yml six-package command. Producer raw run record confirms that exact argv exit 0 in 25.543s on rc.14. Independently reran pinned implementation_coverage.py with pinned ledger on both its exact raw JSON and downloaded hosted candidate JSON: exit 0, 8/8 claims each. |
| Switches / hygiene | Writer false; SPEC_PIN remains 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065. git diff --exit-code for ci.yml, internal/hashing and LOGBOOK against base exits 0. Whitespace check exits 0. One changelog line retains trunk fixes. |
| Full hosted validation | Candidate 3/3 and default 3/3 jobs pass, Linux/macOS race and lint pass, all on matching tree. No full local replay claimed. |

## Executed versus reused evidence

Reviewer ran read-only Git/gh queries, downloaded CI artifacts, independently counted six corpus families, checked protected paths/whitespace, and ran the pinned Python consumption checker twice successfully. Those successful verification calls exited 0. An initial checker invocation without its ledger failed with “coverage ledger not found”; corrected by explicitly supplying the immutable spec ledger. It is not counted as a pass. Some exploratory board queries used unsupported projections and were corrected; no missing read was treated as absence.

Reviewer ran no Go build/test/vet/run, no local lint and no mutation tests. Thus no new reviewer Go duration/crash pair is claimed. Accepted producer rev3 records show GOFLAGS=-work, shared lock, syspolicyd running 369→369 on all recorded calls: candidate targeted 14.751s exit 0, default 13.754s exit 0, snapshot/install 65.858s exit 0, exact spec 25.543s exit 0, build 1.772s exit 0. Earlier twice-stalled marker/scriptworker wrappers exited 130; underlying numeric Go exits unknown. No retry performed. Hosted validation supplies full-matrix/lint evidence per binding decision.

Evidence: https://github.com/relux-works/curator/actions/runs/37017428885 ; existing TASK-261002-2ipeqa_revision3-evidence.tar.gz and TASK-261002-2ipeqa_revision3-results.md. Downloaded hosted candidate stream and exact producer stream both passed the pinned 8-claim checker.

Queried spawn goal before verdict: not goal-bound. Sole branch is changes_requested/to-dev. LOGBOOK unchanged; findings persisted in this artifact and board notes.
