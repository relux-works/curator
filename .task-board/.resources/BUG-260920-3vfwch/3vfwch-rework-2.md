# BUG-260920-3vfwch rework 2 (orchestrator, binding)

Verdict rev2: CHANGES REQUESTED with ONE finding (BUG-260920-3vfwch_review-verdict-rev2.md,
RUN-260920-* opus). F1/F2/N1–N3 are verified resolved — do not touch them.

F3 (must) — `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES`
(`internal/install/gitfixture_retry_test.go:296-333`, atomicity copy same lines) flips the
symlink from a goroutine on a 50 ms timer, so under fork contention the first spawn lands after
the flip and the row fails with `logs = []` (reviewer stress harness: 4 failures / 80
iterations). Make the flip EVENT-DRIVEN: the recording TB performs the symlink swap when it
sees the `retrying once` line (emitted synchronously after attempt 1 failed and before
`sleep(gitFixtureBackoff)`, gitfixture_test.go:127); drop the goroutine and the timer; let the
swap's `os.Symlink`/`os.Rename` errors fail the test (no `_ =`). The reviewer's reference
implementation is attached to the verdict as `zz_fixshape_test.go`
(`TestZZReviewFixOuterEntryEventDrivenFlip`, 70/70 under the stress harness) — adopt that shape
in both copies and keep them byte-identical modulo the package clause/header. Confirm mutants
C (backoff 5 ms) and W (wiring dropped) still die.

results.md: append "Revision 3" with the row shape + why, and, if you rerun the reviewer's
stress harness (`zz_stress_test.go` attached to the verdict), its result. Continue from the
revision-2 tree in the Story workspace (no checkout/clean/stash). Publish revision 3 only when
the configured gate is green.
