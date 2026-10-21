## ungated.log

```text
=== RUN   TestVendorIntakeBoundary
=== RUN   TestVendorIntakeBoundary/unset
=== RUN   TestVendorIntakeBoundary/wrong-digest
--- PASS: TestVendorIntakeBoundary (0.02s)
    --- PASS: TestVendorIntakeBoundary/unset (0.01s)
    --- PASS: TestVendorIntakeBoundary/wrong-digest (0.01s)
=== RUN   TestMuse1410Identity
    intake_test.go:45: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410Identity (0.00s)
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse1410OfflineBoundary
    replay_test.go:365: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410OfflineBoundary (0.00s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.04s)
    --- PASS: TestFrozenSnapshot/regular (0.01s)
    --- PASS: TestFrozenSnapshot/altered (0.00s)
    --- PASS: TestFrozenSnapshot/extra (0.00s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.02s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.01s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.07s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.03s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.03s)
=== RUN   TestReferenceSnapshotPhysicalClosure
    replay_test.go:617: set MUSE1410_SPEC_PYTHON and MUSE1410_SPEC for the published reference CLI
--- SKIP: TestReferenceSnapshotPhysicalClosure (0.00s)
=== RUN   TestMuse1410ParserRefusals
    replay_test.go:738: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410ParserRefusals (0.00s)
=== RUN   TestMuse1410LoaderMatrix
    replay_test.go:760: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410LoaderMatrix (0.00s)
=== RUN   TestMuse1410InteractiveExclusionRegression
    replay_test.go:802: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410InteractiveExclusionRegression (0.00s)
=== RUN   TestMuse1410WorkspaceResidual
    replay_test.go:828: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410WorkspaceResidual (0.00s)
=== RUN   TestMuse1410ExecInteractiveKeyObservation
    replay_test.go:863: vendor replay requires MUSE1410_BINARY, MUSE1410_CURATOR and MUSE1410_RUN_ROOT; no qualification from a skip
--- SKIP: TestMuse1410ExecInteractiveKeyObservation (0.00s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-event
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
PASS
ok  	github.com/relux-works/curator/.research/muse1410	0.398s

```

## full.log

```text
=== RUN   TestVendorIntakeBoundary
=== RUN   TestVendorIntakeBoundary/unset
=== RUN   TestVendorIntakeBoundary/wrong-digest
--- PASS: TestVendorIntakeBoundary (0.02s)
    --- PASS: TestVendorIntakeBoundary/unset (0.02s)
    --- PASS: TestVendorIntakeBoundary/wrong-digest (0.01s)
=== RUN   TestMuse1410Identity
    intake_test.go:58: version exit=0; Muse Code 1.4.10 (1.4.10-R6427.1); exact digest verified
--- PASS: TestMuse1410Identity (0.24s)
=== RUN   TestObservationChecker
--- PASS: TestObservationChecker (0.00s)
=== RUN   TestMuse1410OfflineBoundary
--- PASS: TestMuse1410OfflineBoundary (0.23s)
=== RUN   TestFrozenSnapshot
=== RUN   TestFrozenSnapshot/regular
=== RUN   TestFrozenSnapshot/altered
=== RUN   TestFrozenSnapshot/extra
=== RUN   TestFrozenSnapshot/listed-symlink
=== RUN   TestFrozenSnapshot/unlisted-dangling-link
=== RUN   TestFrozenSnapshot/missing
--- PASS: TestFrozenSnapshot (0.04s)
    --- PASS: TestFrozenSnapshot/regular (0.01s)
    --- PASS: TestFrozenSnapshot/altered (0.01s)
    --- PASS: TestFrozenSnapshot/extra (0.02s)
    --- PASS: TestFrozenSnapshot/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshot/unlisted-dangling-link (0.00s)
    --- PASS: TestFrozenSnapshot/missing (0.00s)
=== RUN   TestFrozenSnapshotSourceInputs
=== RUN   TestFrozenSnapshotSourceInputs/regular
=== RUN   TestFrozenSnapshotSourceInputs/listed-symlink
=== RUN   TestFrozenSnapshotSourceInputs/unlisted-dangling-link
--- PASS: TestFrozenSnapshotSourceInputs (0.04s)
    --- PASS: TestFrozenSnapshotSourceInputs/regular (0.02s)
    --- PASS: TestFrozenSnapshotSourceInputs/listed-symlink (0.01s)
    --- PASS: TestFrozenSnapshotSourceInputs/unlisted-dangling-link (0.01s)
=== RUN   TestReferenceSnapshotPhysicalClosure
=== RUN   TestReferenceSnapshotPhysicalClosure/regular
=== RUN   TestReferenceSnapshotPhysicalClosure/listed-symlink
=== RUN   TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link
--- PASS: TestReferenceSnapshotPhysicalClosure (1.19s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/regular (0.68s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/listed-symlink (0.26s)
    --- PASS: TestReferenceSnapshotPhysicalClosure/unlisted-dangling-link (0.25s)
=== RUN   TestMuse1410ParserRefusals
=== RUN   TestMuse1410ParserRefusals/interactive-exec-flag
=== RUN   TestMuse1410ParserRefusals/interactive-string-false
--- PASS: TestMuse1410ParserRefusals (0.34s)
    --- PASS: TestMuse1410ParserRefusals/interactive-exec-flag (0.09s)
    --- PASS: TestMuse1410ParserRefusals/interactive-string-false (0.08s)
=== RUN   TestMuse1410LoaderMatrix
=== RUN   TestMuse1410LoaderMatrix/exec/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/absent/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/absent/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/absent/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/absent/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=1 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/absent/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/absent/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=0 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/linked/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/linked/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/linked/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/linked/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/linked/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/snapshot/claude-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/snapshot/claude-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/snapshot/codex-rules/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/snapshot/codex-rules/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=0; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/snapshot/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/interactive/snapshot/skills/exclude-true
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
--- PASS: TestMuse1410LoaderMatrix (58.54s)
    --- PASS: TestMuse1410LoaderMatrix/exec/absent/claude-rules/exclude-false (1.26s)
    --- PASS: TestMuse1410LoaderMatrix/exec/absent/claude-rules/exclude-true (1.09s)
    --- PASS: TestMuse1410LoaderMatrix/exec/absent/codex-rules/exclude-false (1.10s)
    --- PASS: TestMuse1410LoaderMatrix/exec/absent/codex-rules/exclude-true (1.14s)
    --- PASS: TestMuse1410LoaderMatrix/exec/absent/skills/exclude-false (1.04s)
    --- PASS: TestMuse1410LoaderMatrix/exec/absent/skills/exclude-true (1.24s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/claude-rules/exclude-false (1.18s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/claude-rules/exclude-true (1.24s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/codex-rules/exclude-false (1.05s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/codex-rules/exclude-true (1.15s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/skills/exclude-false (1.16s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/skills/exclude-true (1.11s)
    --- PASS: TestMuse1410LoaderMatrix/exec/snapshot/claude-rules/exclude-false (1.14s)
    --- PASS: TestMuse1410LoaderMatrix/exec/snapshot/claude-rules/exclude-true (1.12s)
    --- PASS: TestMuse1410LoaderMatrix/exec/snapshot/codex-rules/exclude-false (0.98s)
    --- PASS: TestMuse1410LoaderMatrix/exec/snapshot/codex-rules/exclude-true (1.07s)
    --- PASS: TestMuse1410LoaderMatrix/exec/snapshot/skills/exclude-false (1.04s)
    --- PASS: TestMuse1410LoaderMatrix/exec/snapshot/skills/exclude-true (1.32s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/absent/claude-rules/exclude-false (2.18s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/absent/claude-rules/exclude-true (1.97s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/absent/codex-rules/exclude-false (1.97s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/absent/codex-rules/exclude-true (1.97s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/absent/skills/exclude-false (1.97s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/absent/skills/exclude-true (2.11s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/linked/claude-rules/exclude-false (2.18s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/linked/claude-rules/exclude-true (1.98s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/linked/codex-rules/exclude-false (2.61s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/linked/codex-rules/exclude-true (1.88s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/linked/skills/exclude-false (1.88s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/linked/skills/exclude-true (2.09s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/snapshot/claude-rules/exclude-false (1.97s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/snapshot/claude-rules/exclude-true (2.02s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/snapshot/codex-rules/exclude-false (2.21s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/snapshot/codex-rules/exclude-true (1.86s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/snapshot/skills/exclude-false (2.10s)
    --- PASS: TestMuse1410LoaderMatrix/interactive/snapshot/skills/exclude-true (1.99s)
=== RUN   TestMuse1410InteractiveExclusionRegression
=== RUN   TestMuse1410InteractiveExclusionRegression/linked
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.10-R6427.1.md
=== RUN   TestMuse1410InteractiveExclusionRegression/snapshot
    replay_test.go:820: vendor exit=0; required interactive exclusion FAILED (user_skills=1); see docs/qualification/muse-1.4.10-R6427.1.md
--- PASS: TestMuse1410InteractiveExclusionRegression (4.68s)
    --- PASS: TestMuse1410InteractiveExclusionRegression/linked (1.98s)
    --- PASS: TestMuse1410InteractiveExclusionRegression/snapshot (1.97s)
=== RUN   TestMuse1410WorkspaceResidual
=== RUN   TestMuse1410WorkspaceResidual/exec/linked
=== RUN   TestMuse1410WorkspaceResidual/exec/snapshot
=== RUN   TestMuse1410WorkspaceResidual/interactive/linked
=== RUN   TestMuse1410WorkspaceResidual/interactive/snapshot
--- PASS: TestMuse1410WorkspaceResidual (7.27s)
    --- PASS: TestMuse1410WorkspaceResidual/exec/linked (1.00s)
    --- PASS: TestMuse1410WorkspaceResidual/exec/snapshot (1.29s)
    --- PASS: TestMuse1410WorkspaceResidual/interactive/linked (2.07s)
    --- PASS: TestMuse1410WorkspaceResidual/interactive/snapshot (2.13s)
=== RUN   TestMuse1410ExecInteractiveKeyObservation
    replay_test.go:871: vendor exit=0; exec accepts -c context.foreign_personal_skills=false; exclusion semantics not measured by this parser case
--- PASS: TestMuse1410ExecInteractiveKeyObservation (1.58s)
=== RUN   TestReviewContradictoryObservation
=== RUN   TestReviewContradictoryObservation/count-one-then-zero
=== RUN   TestReviewContradictoryObservation/outcome-failed-then-ready
=== RUN   TestReviewContradictoryObservation/duplicate-count
=== RUN   TestReviewContradictoryObservation/duplicate-event
=== RUN   TestReviewContradictoryObservation/quoted-decoy
=== RUN   TestReviewContradictoryObservation/malformed-count
=== RUN   TestReviewContradictoryObservation/outcome-ready-then-failed
=== RUN   TestReviewContradictoryObservation/duplicate-outcome
=== RUN   TestReviewContradictoryObservation/overflow-count
=== RUN   TestReviewContradictoryObservation/count-zero-then-one
--- PASS: TestReviewContradictoryObservation (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-one-then-zero (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-failed-then-ready (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-event (0.00s)
    --- PASS: TestReviewContradictoryObservation/quoted-decoy (0.00s)
    --- PASS: TestReviewContradictoryObservation/malformed-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/outcome-ready-then-failed (0.00s)
    --- PASS: TestReviewContradictoryObservation/duplicate-outcome (0.00s)
    --- PASS: TestReviewContradictoryObservation/overflow-count (0.00s)
    --- PASS: TestReviewContradictoryObservation/count-zero-then-one (0.00s)
PASS
ok  	github.com/relux-works/curator/.research/muse1410	74.380s

```

## plant.log

```text
=== RUN   TestMuse1410LoaderMatrix
=== RUN   TestMuse1410LoaderMatrix/exec/linked/skills/exclude-false
    replay_test.go:790: vendor exit=0; user_sources=1 foreign_sources=0 user_skills=1; no model claim
=== RUN   TestMuse1410LoaderMatrix/exec/linked/skills/exclude-true
    replay_test.go:788: skills.catalog_load observed user_skills=1; want 0
--- FAIL: TestMuse1410LoaderMatrix (3.27s)
    --- PASS: TestMuse1410LoaderMatrix/exec/linked/skills/exclude-false (1.09s)
    --- FAIL: TestMuse1410LoaderMatrix/exec/linked/skills/exclude-true (1.40s)
FAIL
FAIL	github.com/relux-works/curator/.research/muse1410	3.448s
FAIL

```

## mutant.log

```text
=== RUN   TestObservationChecker
    replay_test.go:121: checker admitted one foreign skill
--- FAIL: TestObservationChecker (0.00s)
FAIL
FAIL	github.com/relux-works/curator/.temp/review-ujp9nh/mutant	0.285s
FAIL

```

## vet.log

```text

```

## hygiene.log

```text
Scope, README/CHANGELOG, personal-path screening, source digests PASS. Initial broad /home/ search matched relative golden/home fixtures only.

```

## hosted.json

```text
{"conclusion":"success","head_sha":"e88fdf2dbec7f94c48b0232a350a0a9f9484fdeb","status":"completed"}

```
