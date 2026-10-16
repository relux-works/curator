# Review execution evidence — TASK-261016-i5gwaf

Private locators redacted. Commands and exit interpretations are in the verdict.

## identity.json

```text
{
  "candidate_tree": "187619f43617eeaec4d7cac2df091199581c9c1b",
  "base": "dccac3199aa5c543c589c129f5ef2bc437523a2a",
  "upstream": "58b2aac337f73fb313893611d9de28e9a4fcff0a",
  "overlap": [],
  "candidate_files": 69,
  "hosted_snapshot": "943be40892ff2253aa4ca0c71a6e34aa47287894",
  "normative_main": "15d970111277b1f68b3f7a4ddd63bc339b2c0a3f",
  "probe_sha256": "78ecf817e292c2a7032bea08cf403dc9f2e86bd015eefdb3caf82abd02f94da1"
}

```

## sweep.md

```text
# Review sweep — TASK-261016-i5gwaf

Candidate revision 3, tree 187619f43617eeaec4d7cac2df091199581c9c1b.

No catalog-keyed table or budget was supplied; use the supplied single surface. AC coverage checked before implementation: 6 of 6 rows have named driving tests.

| Surface | Result | Executed attacks |
|---|---|---|
| Vendor-build evidence authorizing form-specific exclusion | held | Prior contradictory count/outcome/event regressions; actual-source listed and dangling symlink attacks; real reference CLI; missing/failed/truncated records; altered/extra/missing fixture bytes. Full real vendor replay passes as an observation witness. Seven narrowing mutants fail at intended assertions. Held covers only these named attacks. |

Free hunt follows this persisted sweep. Final gate and remaining controls will be reported separately.

Additional controls completed: offline-loopback narrowing mutant exit 1 at network deny assertion; exec omission exit 1 with no-exclusion controls passing on both roots; interactive omission exit 0 (expected survivor, vendor failure). Free hunt reference CLI rejected directory, metadata and bundle-root symlinks, missing and changed root bytes and an extra file; valid regular control passed and source bytes were unchanged. No blocking free-hunt finding.

```

## focused.log

```text
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.01s)
    --- PASS: TestFrozenSnapshot/regular (0.00s)
    --- PASS: TestFrozenSnapshot/altered (0.00s)
    --- PASS: TestFrozenSnapshot/extra (0.00s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.00s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.00s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.03s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- PASS: TestReferenceSnapshotPhysicalClosure (1.09s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.60s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.24s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.25s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/quoted-decoy
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	1.444s

```

## replay.log

```text
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse148OfflineBoundary
--- PASS: TestMuse148OfflineBoundary (0.26s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.01s)
    --- PASS: TestFrozenSnapshot/regular (0.00s)
    --- PASS: TestFrozenSnapshot/altered (0.00s)
    --- PASS: TestFrozenSnapshot/extra (0.00s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.00s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.00s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.02s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- PASS: TestReferenceSnapshotPhysicalClosure (1.03s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.55s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.24s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.23s)
=== RUN   TestMuse148ParserRefusals
=== RUN   TestMuse148ParserRefusals/interactive-exec-flag
=== RUN   TestMuse148ParserRefusals/interactive-string-false
--- PASS: TestMuse148ParserRefusals (0.37s)
    --- PASS: TestMuse148ParserRefusals/interactive-exec-flag (0.13s)
    --- PASS: TestMuse148ParserRefusals/interactive-string-false (0.06s)
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (56.32s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-false (0.97s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/claude-rules/exclude-true (0.94s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-false (0.90s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/codex-rules/exclude-true (0.99s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-false (1.20s)
    --- PASS: TestMuse148LoaderMatrix/exec/absent/skills/exclude-true (1.31s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-false (0.98s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/claude-rules/exclude-true (0.92s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-false (0.89s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/codex-rules/exclude-true (0.95s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (1.10s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (1.26s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-false (1.31s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/claude-rules/exclude-true (1.00s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-false (0.92s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/codex-rules/exclude-true (0.91s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (0.94s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (0.96s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-false (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/claude-rules/exclude-true (2.12s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-false (1.76s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/codex-rules/exclude-true (2.09s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-false (2.22s)
    --- PASS: TestMuse148LoaderMatrix/interactive/absent/skills/exclude-true (2.32s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-false (2.56s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/claude-rules/exclude-true (2.35s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-false (1.97s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/codex-rules/exclude-true (1.96s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (2.08s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (2.07s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (1.87s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (1.87s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (1.86s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (1.88s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (1.98s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (1.65s)
=== RUN   TestMuse148InteractiveExclusionRegression
=== RUN   TestMuse148InteractiveExclusionRegression/linked
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
=== RUN   TestMuse148InteractiveExclusionRegression/snapshot
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.8-R6256.1.md
--- PASS: TestMuse148InteractiveExclusionRegression (4.32s)
    --- PASS: TestMuse148InteractiveExclusionRegression/linked (1.66s)
    --- PASS: TestMuse148InteractiveExclusionRegression/snapshot (1.96s)
=== RUN   TestMuse148WorkspaceResidual
=== RUN   TestMuse148WorkspaceResidual/exec/linked
=== RUN   TestMuse148WorkspaceResidual/exec/snapshot
=== RUN   TestMuse148WorkspaceResidual/interactive/linked
=== RUN   TestMuse148WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse148WorkspaceResidual (6.34s)
    --- PASS: TestMuse148WorkspaceResidual/exec/linked (0.93s)
    --- PASS: TestMuse148WorkspaceResidual/exec/snapshot (0.87s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/linked (1.87s)
    --- PASS: TestMuse148WorkspaceResidual/interactive/snapshot (1.87s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	69.024s

```

## checker-duplicate-count.log

```text
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="ready" user_skills=0 user_skills=1
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
--- FAIL: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- FAIL: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/checker-duplicate-count	0.352s
FAIL

```

## checker-duplicate-outcome.log

```text
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
    review_attack_test.go:26: contradictory record admitted: event="skills.catalog_load" outcome="failed" outcome="ready" user_skills=0
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
--- FAIL: TestReviewContradictoryObservation (0.00s)
    --- FAIL: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/checker-duplicate-outcome	0.382s
FAIL

```

## snapshot-go-listed-link.log

```text
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
    replay_test.go:498: listed-symlink snapshot admitted
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- FAIL: TestFrozenSnapshot (0.06s)
    --- PASS: TestFrozenSnapshot/regular (0.02s)
    --- PASS: TestFrozenSnapshot/altered (0.01s)
    --- PASS: TestFrozenSnapshot/extra (0.00s)
    --- FAIL: TestFrozenSnapshot/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.01s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
    replay_test.go:540: listed-symlink source fixture exit=0: === RUN   TestFrozenSnapshot
        === RUN   TestFrozenSnapshot/regular
        --- PASS: TestFrozenSnapshot (0.00s)
            --- PASS: TestFrozenSnapshot/regular (0.00s)
        PASS
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- FAIL: TestFrozenSnapshotSourceInputs (0.04s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.02s)
    --- FAIL: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/snapshot-go-listed-link	0.386s
FAIL

```

## snapshot-reference-listed-link.log

```text
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
    replay_test.go:632: listed-symlink reference exit=0: Reference check_bundle accepted exact frozen metadata and complete regular-file closure
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- FAIL: TestReferenceSnapshotPhysicalClosure (1.39s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.56s)
    --- FAIL: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.57s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.27s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/snapshot-reference-listed-link	1.767s
FAIL

```

## snapshot-reference-dangling-link.log

```text
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
    replay_test.go:632: unlisted-dangling-link reference exit=0: Reference check_bundle accepted exact frozen metadata and complete regular-file closure
--- FAIL: TestReferenceSnapshotPhysicalClosure (1.35s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.56s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.24s)
    --- FAIL: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.54s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/snapshot-reference-dangling-link	1.654s
FAIL

```

## checker-one-skill.log

```text
=== RUN   TestObservationChecker
    replay_test.go:121: checker admitted one foreign skill
--- FAIL: TestObservationChecker (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/checker-one-skill	0.414s
FAIL

```

## snapshot-extra-entry.log

```text
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
    replay_test.go:498: extra snapshot admitted
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- FAIL: TestFrozenSnapshot (0.01s)
    --- PASS: TestFrozenSnapshot/regular (0.00s)
    --- PASS: TestFrozenSnapshot/altered (0.00s)
    --- FAIL: TestFrozenSnapshot/extra (0.00s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.00s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.00s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/snapshot-extra-entry	0.339s
FAIL

```

## offline-mutant.log

```text
=== RUN   TestMuse148OfflineBoundary
    replay_test.go:379: network deny exit=0: MUSE148_NETWORK_CONTROL
--- FAIL: TestMuse148OfflineBoundary (0.21s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-muse-r3/offline-loopback	0.608s
FAIL

```

## omit-exec.log

```text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse148LoaderMatrix (4.58s)
    --- PASS: TestMuse148LoaderMatrix/exec/linked/skills/exclude-false (0.95s)
    --- FAIL: TestMuse148LoaderMatrix/exec/linked/skills/exclude-true (0.98s)
    --- PASS: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-false (0.95s)
    --- FAIL: TestMuse148LoaderMatrix/exec/snapshot/skills/exclude-true (0.93s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse148	4.869s
FAIL

```

## omit-interactive.log

```text
=== RUN   TestMuse148LoaderMatrix
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse148LoaderMatrix (7.91s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-false (1.64s)
    --- PASS: TestMuse148LoaderMatrix/interactive/linked/skills/exclude-true (1.86s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-false (1.86s)
    --- PASS: TestMuse148LoaderMatrix/interactive/snapshot/skills/exclude-true (1.85s)
PASS
ok  	github.com/relux-works/curator/.research/muse148	8.368s

```

## free-hunt.log

```text
regular 0 accepted
directory-link 1 ['ValueError: physical-entry-type']
metadata-link 1 ['ValueError: physical-entry-type']
root-link 1 ['ValueError: physical-entry-type']
missing-root 1 ['ValueError: entry-set']
changed-root 1 ['ValueError: file-identity']
extra-file 1 ['ValueError: entry-set']

```

## lint.log

```text
0 issues.

```

## producer-hosted.json

```text
{"attempt":2,"conclusion":"success","headSha":"a88ce7997152b97e99389b6bbf9303e5e210e75f","jobs":[{"completedAt":"2026-10-16T09:23:17Z","conclusion":"success","databaseId":116886088789,"name":"Test (windows-latest)","startedAt":"2026-10-16T08:33:28Z","status":"completed","steps":[{"completedAt":"2026-10-16T08:33:34Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T08:33:29Z","status":"completed"},{"completedAt":"2026-10-16T08:33:58Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T08:33:34Z","status":"completed"},{"completedAt":"2026-10-16T08:34:02Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T08:33:58Z","status":"completed"},{"completedAt":"2026-10-16T08:34:39Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T08:34:02Z","status":"completed"},{"completedAt":"2026-10-16T08:34:44Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T08:34:39Z","status":"completed"},{"completedAt":"2026-10-16T08:34:44Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T08:34:44Z","status":"completed"},{"completedAt":"2026-10-16T08:34:50Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T08:34:44Z","status":"completed"},{"completedAt":"2026-10-16T08:34:51Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T08:34:50Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T08:34:51Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:35:01Z","conclusion":"skipped","name":"gofmt check","number":12,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:35:20Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T08:35:01Z","status":"completed"},{"completedAt":"2026-10-16T08:36:07Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T08:35:20Z","status":"completed"},{"completedAt":"2026-10-16T09:23:08Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T08:36:07Z","status":"completed"},{"completedAt":"2026-10-16T09:23:09Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T09:23:08Z","status":"completed"},{"completedAt":"2026-10-16T09:23:10Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T09:23:09Z","status":"completed"},{"completedAt":"2026-10-16T09:23:10Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T09:23:10Z","status":"completed"},{"completedAt":"2026-10-16T09:23:12Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T09:23:10Z","status":"completed"},{"completedAt":"2026-10-16T09:23:15Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T09:23:12Z","status":"completed"},{"completedAt":"2026-10-16T09:23:15Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T09:23:15Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886088789"},{"completedAt":"2026-10-16T07:49:50Z","conclusion":"skipped","databaseId":116886089253,"name":"Test (rose-air)","startedAt":"2026-10-16T08:33:24Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886089253"},{"completedAt":"2026-10-16T07:50:41Z","conclusion":"success","databaseId":116886089756,"name":"Go 1.25.5 driver (ubuntu-latest)","startedAt":"2026-10-16T07:49:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:04Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:55Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:17Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:17Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:17Z","status":"completed"},{"completedAt":"2026-10-16T07:50:22Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:17Z","status":"completed"},{"completedAt":"2026-10-16T07:50:24Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:50:22Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:50:24Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:50:38Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886089756"},{"completedAt":"2026-10-16T07:49:50Z","conclusion":"skipped","databaseId":116886090010,"name":"Candidate suite (${{ matrix.os }})","startedAt":"2026-10-16T08:33:24Z","status":"completed","steps":[],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090010"},{"completedAt":"2026-10-16T07:51:33Z","conclusion":"success","databaseId":116886090219,"name":"Gate self-test (windows-latest)","startedAt":"2026-10-16T07:50:22Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:23Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:22Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:23Z","status":"completed"},{"completedAt":"2026-10-16T07:51:29Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T07:50:38Z","status":"completed"},{"completedAt":"2026-10-16T07:51:31Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:51:29Z","status":"completed"},{"completedAt":"2026-10-16T07:51:31Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:51:31Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090219"},{"completedAt":"2026-10-16T07:53:52Z","conclusion":"success","databaseId":116886090295,"name":"Go 1.27.1 driver (macos-latest)","startedAt":"2026-10-16T07:51:09Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:51:11Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:51:09Z","status":"completed"},{"completedAt":"2026-10-16T07:51:21Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:51:11Z","status":"completed"},{"completedAt":"2026-10-16T07:51:22Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:51:21Z","status":"completed"},{"completedAt":"2026-10-16T07:51:48Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:51:22Z","status":"completed"},{"completedAt":"2026-10-16T07:51:48Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:51:48Z","status":"completed"},{"completedAt":"2026-10-16T07:53:27Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:51:48Z","status":"completed"},{"completedAt":"2026-10-16T07:53:29Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:53:27Z","status":"completed"},{"completedAt":"2026-10-16T07:53:44Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:53:29Z","status":"completed"},{"completedAt":"2026-10-16T07:53:45Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:44Z","status":"completed"},{"completedAt":"2026-10-16T07:53:45Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:45Z","status":"completed"},{"completedAt":"2026-10-16T07:53:46Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:53:45Z","status":"completed"},{"completedAt":"2026-10-16T07:53:47Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:53:46Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090295"},{"completedAt":"2026-10-16T08:14:55Z","conclusion":"success","databaseId":116886090346,"name":"Go 1.25.5 driver (macos-latest)","startedAt":"2026-10-16T08:13:10Z","status":"completed","steps":[{"completedAt":"2026-10-16T08:13:11Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T08:13:10Z","status":"completed"},{"completedAt":"2026-10-16T08:13:20Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T08:13:11Z","status":"completed"},{"completedAt":"2026-10-16T08:13:21Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T08:13:20Z","status":"completed"},{"completedAt":"2026-10-16T08:13:40Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T08:13:21Z","status":"completed"},{"completedAt":"2026-10-16T08:13:41Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T08:13:40Z","status":"completed"},{"completedAt":"2026-10-16T08:14:39Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T08:13:41Z","status":"completed"},{"completedAt":"2026-10-16T08:14:40Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T08:14:39Z","status":"completed"},{"completedAt":"2026-10-16T08:14:50Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T08:14:40Z","status":"completed"},{"completedAt":"2026-10-16T08:14:50Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T08:14:50Z","status":"completed"},{"completedAt":"2026-10-16T08:14:51Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T08:14:50Z","status":"completed"},{"completedAt":"2026-10-16T08:14:51Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T08:14:51Z","status":"completed"},{"completedAt":"2026-10-16T08:14:52Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T08:14:51Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090346"},{"completedAt":"2026-10-16T07:53:33Z","conclusion":"success","databaseId":116886090419,"name":"Go 1.25.5 driver (windows-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:11Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:11Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:53:06Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:53:10Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:53:06Z","status":"completed"},{"completedAt":"2026-10-16T07:53:26Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:53:10Z","status":"completed"},{"completedAt":"2026-10-16T07:53:26Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:26Z","status":"completed"},{"completedAt":"2026-10-16T07:53:29Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:26Z","status":"completed"},{"completedAt":"2026-10-16T07:53:31Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:53:29Z","status":"completed"},{"completedAt":"2026-10-16T07:53:31Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:53:31Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090419"},{"completedAt":"2026-10-16T07:50:31Z","conclusion":"success","databaseId":116886090432,"name":"Go 1.27.1 driver (ubuntu-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:00Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:02Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:00Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:02Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:17Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:19Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:50:17Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:50:19Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:50:28Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090432"},{"completedAt":"2026-10-16T08:10:32Z","conclusion":"success","databaseId":116886090523,"name":"Test (ubuntu-latest)","startedAt":"2026-10-16T07:49:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:03Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:55Z","status":"completed"},{"completedAt":"2026-10-16T07:50:04Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:03Z","status":"completed"},{"completedAt":"2026-10-16T07:50:12Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:50:12Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:40Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T08:10:27Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T07:50:40Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T08:10:27Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T08:10:29Z","status":"completed"},{"completedAt":"2026-10-16T08:10:29Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T08:10:29Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090523"},{"completedAt":"2026-10-16T08:24:41Z","conclusion":"success","databaseId":116886090682,"name":"Race (macos-latest)","startedAt":"2026-10-16T07:49:58Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:00Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:58Z","status":"completed"},{"completedAt":"2026-10-16T07:50:14Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:00Z","status":"completed"},{"completedAt":"2026-10-16T07:50:16Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:14Z","status":"completed"},{"completedAt":"2026-10-16T07:50:52Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:16Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:50:52Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:55Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:55Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:50:55Z","status":"completed"},{"completedAt":"2026-10-16T07:51:02Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:50:55Z","status":"completed"},{"completedAt":"2026-10-16T07:51:03Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:51:02Z","status":"completed"},{"completedAt":"2026-10-16T08:24:27Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T07:51:03Z","status":"completed"},{"completedAt":"2026-10-16T08:24:29Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T08:24:27Z","status":"completed"},{"completedAt":"2026-10-16T08:24:29Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T08:24:29Z","status":"completed"},{"completedAt":"2026-10-16T08:24:30Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T08:24:29Z","status":"completed"},{"completedAt":"2026-10-16T08:24:32Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T08:24:30Z","status":"completed"},{"completedAt":"2026-10-16T08:24:33Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T08:24:32Z","status":"completed"},{"completedAt":"2026-10-16T08:24:38Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T08:24:33Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886090682"},{"completedAt":"2026-10-16T07:54:06Z","conclusion":"success","databaseId":116886091052,"name":"Go 1.27.1 driver (windows-latest)","startedAt":"2026-10-16T07:50:32Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:34Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:33Z","status":"completed"},{"completedAt":"2026-10-16T07:50:45Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:34Z","status":"completed"},{"completedAt":"2026-10-16T07:50:48Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:45Z","status":"completed"},{"completedAt":"2026-10-16T07:51:18Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:48Z","status":"completed"},{"completedAt":"2026-10-16T07:51:19Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:51:18Z","status":"completed"},{"completedAt":"2026-10-16T07:53:38Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:51:19Z","status":"completed"},{"completedAt":"2026-10-16T07:53:41Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:53:38Z","status":"completed"},{"completedAt":"2026-10-16T07:53:59Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:53:41Z","status":"completed"},{"completedAt":"2026-10-16T07:53:59Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:59Z","status":"completed"},{"completedAt":"2026-10-16T07:54:01Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:59Z","status":"completed"},{"completedAt":"2026-10-16T07:54:04Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:54:01Z","status":"completed"},{"completedAt":"2026-10-16T07:54:04Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:54:04Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886091052"},{"completedAt":"2026-10-16T07:53:11Z","conclusion":"success","databaseId":116886091165,"name":"Go 1.26.0 driver (windows-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:09Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:09Z","status":"completed"},{"completedAt":"2026-10-16T07:50:37Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:52:43Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:37Z","status":"completed"},{"completedAt":"2026-10-16T07:52:47Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:52:43Z","status":"completed"},{"completedAt":"2026-10-16T07:53:04Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:52:47Z","status":"completed"},{"completedAt":"2026-10-16T07:53:04Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:53:04Z","status":"completed"},{"completedAt":"2026-10-16T07:53:07Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:53:04Z","status":"completed"},{"completedAt":"2026-10-16T07:53:09Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:53:07Z","status":"completed"},{"completedAt":"2026-10-16T07:53:09Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:53:09Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886091165"},{"completedAt":"2026-10-16T08:20:48Z","conclusion":"success","databaseId":116886123572,"name":"Race (ubuntu-latest)","startedAt":"2026-10-16T07:50:09Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:10Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:09Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:10Z","status":"completed"},{"completedAt":"2026-10-16T07:50:19Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:18Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:19Z","status":"completed"},{"completedAt":"2026-10-16T07:50:29Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:29Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:50:29Z","status":"completed"},{"completedAt":"2026-10-16T07:50:30Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:50:29Z","status":"completed"},{"completedAt":"2026-10-16T07:50:31Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:50:30Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:50:31Z","status":"completed"},{"completedAt":"2026-10-16T07:50:38Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:50:38Z","status":"completed"},{"completedAt":"2026-10-16T08:20:44Z","conclusion":"success","name":"go test -race + platform-case gate","number":11,"startedAt":"2026-10-16T07:50:38Z","status":"completed"},{"completedAt":"2026-10-16T08:20:45Z","conclusion":"success","name":"Upload race evidence","number":12,"startedAt":"2026-10-16T08:20:44Z","status":"completed"},{"completedAt":"2026-10-16T08:20:45Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":21,"startedAt":"2026-10-16T08:20:45Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":22,"startedAt":"2026-10-16T08:20:45Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":23,"startedAt":"2026-10-16T08:20:46Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":24,"startedAt":"2026-10-16T08:20:46Z","status":"completed"},{"completedAt":"2026-10-16T08:20:46Z","conclusion":"success","name":"Complete job","number":25,"startedAt":"2026-10-16T08:20:46Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886123572"},{"completedAt":"2026-10-16T07:50:20Z","conclusion":"success","databaseId":116886132197,"name":"Gate self-test (ubuntu-latest)","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:01Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T07:50:01Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:50:18Z","status":"completed"},{"completedAt":"2026-10-16T07:50:18Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:50:18Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886132197"},{"completedAt":"2026-10-16T07:51:07Z","conclusion":"success","databaseId":116886133989,"name":"Interop conformance gate","startedAt":"2026-10-16T07:50:47Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:48Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:48Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:48Z","status":"completed"},{"completedAt":"2026-10-16T07:50:53Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:51:02Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:53Z","status":"completed"},{"completedAt":"2026-10-16T07:51:02Z","conclusion":"success","name":"Verify Go toolchain identity","number":5,"startedAt":"2026-10-16T07:51:02Z","status":"completed"},{"completedAt":"2026-10-16T07:51:04Z","conclusion":"success","name":"Shared protocol suite","number":6,"startedAt":"2026-10-16T07:51:02Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":10,"startedAt":"2026-10-16T07:51:04Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":11,"startedAt":"2026-10-16T07:51:05Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":12,"startedAt":"2026-10-16T07:51:05Z","status":"completed"},{"completedAt":"2026-10-16T07:51:05Z","conclusion":"success","name":"Complete job","number":13,"startedAt":"2026-10-16T07:51:05Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886133989"},{"completedAt":"2026-10-16T07:50:46Z","conclusion":"success","databaseId":116886134849,"name":"Go 1.26.0 driver (ubuntu-latest)","startedAt":"2026-10-16T07:50:03Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:08Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:09Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:50:08Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:50:09Z","status":"completed"},{"completedAt":"2026-10-16T07:50:21Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:26Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:50:21Z","status":"completed"},{"completedAt":"2026-10-16T07:50:28Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:50:26Z","status":"completed"},{"completedAt":"2026-10-16T07:50:42Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:50:28Z","status":"completed"},{"completedAt":"2026-10-16T07:50:42Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:50:42Z","status":"completed"},{"completedAt":"2026-10-16T07:50:42Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:50:42Z","status":"completed"},{"completedAt":"2026-10-16T07:50:43Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:50:42Z","status":"completed"},{"completedAt":"2026-10-16T07:50:43Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:50:43Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886134849"},{"completedAt":"2026-10-16T07:50:48Z","conclusion":"success","databaseId":116886136467,"name":"Gate self-test (macos-latest)","startedAt":"2026-10-16T07:50:23Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:50:24Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:50:24Z","status":"completed"},{"completedAt":"2026-10-16T07:50:35Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:50:24Z","status":"completed"},{"completedAt":"2026-10-16T07:50:45Z","conclusion":"success","name":"Self-test the CI gates","number":3,"startedAt":"2026-10-16T07:50:35Z","status":"completed"},{"completedAt":"2026-10-16T07:50:45Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:50:45Z","status":"completed"},{"completedAt":"2026-10-16T07:50:46Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:50:45Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886136467"},{"completedAt":"2026-10-16T07:50:07Z","conclusion":"success","databaseId":116886139040,"name":"Naming gate","startedAt":"2026-10-16T07:49:53Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:54Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:03Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:54Z","status":"completed"},{"completedAt":"2026-10-16T07:50:04Z","conclusion":"success","name":"Employer name gate","number":3,"startedAt":"2026-10-16T07:50:03Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":6,"startedAt":"2026-10-16T07:50:04Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Complete job","number":7,"startedAt":"2026-10-16T07:50:05Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886139040"},{"completedAt":"2026-10-16T07:52:27Z","conclusion":"success","databaseId":116886141826,"name":"Go 1.26.0 driver (macos-latest)","startedAt":"2026-10-16T07:51:00Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:51:01Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:51:00Z","status":"completed"},{"completedAt":"2026-10-16T07:51:09Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:51:01Z","status":"completed"},{"completedAt":"2026-10-16T07:51:10Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:51:09Z","status":"completed"},{"completedAt":"2026-10-16T07:51:21Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:51:10Z","status":"completed"},{"completedAt":"2026-10-16T07:51:21Z","conclusion":"success","name":"Verify family toolchain identity","number":5,"startedAt":"2026-10-16T07:51:21Z","status":"completed"},{"completedAt":"2026-10-16T07:52:15Z","conclusion":"success","name":"Driver qualification and reproducibility","number":6,"startedAt":"2026-10-16T07:51:21Z","status":"completed"},{"completedAt":"2026-10-16T07:52:16Z","conclusion":"success","name":"Driver vet","number":7,"startedAt":"2026-10-16T07:52:15Z","status":"completed"},{"completedAt":"2026-10-16T07:52:21Z","conclusion":"success","name":"Build manager","number":8,"startedAt":"2026-10-16T07:52:16Z","status":"completed"},{"completedAt":"2026-10-16T07:52:22Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":14,"startedAt":"2026-10-16T07:52:21Z","status":"completed"},{"completedAt":"2026-10-16T07:52:22Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":15,"startedAt":"2026-10-16T07:52:22Z","status":"completed"},{"completedAt":"2026-10-16T07:52:22Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":16,"startedAt":"2026-10-16T07:52:22Z","status":"completed"},{"completedAt":"2026-10-16T07:52:23Z","conclusion":"success","name":"Complete job","number":17,"startedAt":"2026-10-16T07:52:22Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886141826"},{"completedAt":"2026-10-16T08:15:27Z","conclusion":"success","databaseId":116886145726,"name":"Test (macos-latest)","startedAt":"2026-10-16T07:52:33Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:52:35Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:52:34Z","status":"completed"},{"completedAt":"2026-10-16T07:52:47Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:52:35Z","status":"completed"},{"completedAt":"2026-10-16T07:52:49Z","conclusion":"success","name":"Checkout authoritative protocol suite","number":3,"startedAt":"2026-10-16T07:52:47Z","status":"completed"},{"completedAt":"2026-10-16T07:53:14Z","conclusion":"success","name":"Run actions/setup-go@v5","number":4,"startedAt":"2026-10-16T07:52:49Z","status":"completed"},{"completedAt":"2026-10-16T07:53:14Z","conclusion":"success","name":"Run actions/setup-node@v4","number":5,"startedAt":"2026-10-16T07:53:14Z","status":"completed"},{"completedAt":"2026-10-16T07:53:14Z","conclusion":"success","name":"Verify pnpm pin matches Go constant","number":6,"startedAt":"2026-10-16T07:53:14Z","status":"completed"},{"completedAt":"2026-10-16T07:53:16Z","conclusion":"success","name":"Install pinned pnpm via npm","number":7,"startedAt":"2026-10-16T07:53:14Z","status":"completed"},{"completedAt":"2026-10-16T07:53:16Z","conclusion":"success","name":"Verify rust pin matches Go constant","number":8,"startedAt":"2026-10-16T07:53:16Z","status":"completed"},{"completedAt":"2026-10-16T07:53:23Z","conclusion":"success","name":"Install pinned Rust toolchain via rustup","number":9,"startedAt":"2026-10-16T07:53:16Z","status":"completed"},{"completedAt":"2026-10-16T07:53:23Z","conclusion":"success","name":"Verify Go toolchain identity","number":10,"startedAt":"2026-10-16T07:53:23Z","status":"completed"},{"completedAt":"2026-10-16T07:53:24Z","conclusion":"success","name":"Verify versioned Go install compatibility","number":11,"startedAt":"2026-10-16T07:53:23Z","status":"completed"},{"completedAt":"2026-10-16T07:53:24Z","conclusion":"success","name":"gofmt check","number":12,"startedAt":"2026-10-16T07:53:24Z","status":"completed"},{"completedAt":"2026-10-16T07:53:36Z","conclusion":"success","name":"go vet","number":13,"startedAt":"2026-10-16T07:53:24Z","status":"completed"},{"completedAt":"2026-10-16T07:53:48Z","conclusion":"success","name":"Ledger consistency","number":14,"startedAt":"2026-10-16T07:53:36Z","status":"completed"},{"completedAt":"2026-10-16T08:15:14Z","conclusion":"success","name":"go test + platform-case gate","number":15,"startedAt":"2026-10-16T07:53:48Z","status":"completed"},{"completedAt":"2026-10-16T08:15:15Z","conclusion":"success","name":"Upload gate evidence","number":16,"startedAt":"2026-10-16T08:15:14Z","status":"completed"},{"completedAt":"2026-10-16T08:15:15Z","conclusion":"success","name":"Post Run actions/setup-node@v4","number":29,"startedAt":"2026-10-16T08:15:15Z","status":"completed"},{"completedAt":"2026-10-16T08:15:16Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":30,"startedAt":"2026-10-16T08:15:15Z","status":"completed"},{"completedAt":"2026-10-16T08:15:17Z","conclusion":"success","name":"Post Checkout authoritative protocol suite","number":31,"startedAt":"2026-10-16T08:15:16Z","status":"completed"},{"completedAt":"2026-10-16T08:15:18Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":32,"startedAt":"2026-10-16T08:15:17Z","status":"completed"},{"completedAt":"2026-10-16T08:15:23Z","conclusion":"success","name":"Complete job","number":33,"startedAt":"2026-10-16T08:15:18Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886145726"},{"completedAt":"2026-10-16T07:50:55Z","conclusion":"success","databaseId":116886149092,"name":"Lint","startedAt":"2026-10-16T07:49:52Z","status":"completed","steps":[{"completedAt":"2026-10-16T07:49:55Z","conclusion":"success","name":"Set up job","number":1,"startedAt":"2026-10-16T07:49:53Z","status":"completed"},{"completedAt":"2026-10-16T07:50:05Z","conclusion":"success","name":"Run actions/checkout@v4","number":2,"startedAt":"2026-10-16T07:49:55Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Run actions/setup-go@v5","number":3,"startedAt":"2026-10-16T07:50:05Z","status":"completed"},{"completedAt":"2026-10-16T07:50:13Z","conclusion":"success","name":"Verify Go toolchain identity","number":4,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:50Z","conclusion":"success","name":"Run golangci/golangci-lint-action@v7","number":5,"startedAt":"2026-10-16T07:50:13Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"No broad suppression","number":6,"startedAt":"2026-10-16T07:50:50Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Verify GoReleaser rc channel values","number":7,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Post Run golangci/golangci-lint-action@v7","number":12,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Post Run actions/setup-go@v5","number":13,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Post Run actions/checkout@v4","number":14,"startedAt":"2026-10-16T07:50:51Z","status":"completed"},{"completedAt":"2026-10-16T07:50:51Z","conclusion":"success","name":"Complete job","number":15,"startedAt":"2026-10-16T07:50:51Z","status":"completed"}],"url":"https://github.com/relux-works/curator/actions/runs/38869688515/job/116886149092"}],"status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38869688515"}

```

## gate.log

```text
remote gate: attempt 1/3: pushing 943be40892ff2253aa4ca0c71a6e34aa47287894 as gate/STORY-261016-5y3h2r/261016-110029-97860-1
remote: 
remote: Create a pull request for 'gate/STORY-261016-5y3h2r/261016-110029-97860-1' on GitHub by visiting:        
remote:      https://github.com/relux-works/curator/pull/new/gate/STORY-261016-5y3h2r/261016-110029-97860-1        
remote: 
remote gate: run 38890939343 (https://github.com/relux-works/curator/actions/runs/38890939343)

```

## duplicate-cancel.log

```text
✓ Request to cancel workflow 38890939343 submitted.
To github.com:relux-works/curator.git
 - [deleted]           gate/STORY-261016-5y3h2r/261016-110029-97860-1

```

## duplicate-terminal.json

```text
{"conclusion":"cancelled","headSha":"943be40892ff2253aa4ca0c71a6e34aa47287894","status":"completed","url":"https://github.com/relux-works/curator/actions/runs/38890939343"}

```
