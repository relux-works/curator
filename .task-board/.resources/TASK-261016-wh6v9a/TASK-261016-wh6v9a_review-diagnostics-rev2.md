# Diagnostic replays of first-attempt Ubuntu failures

These macOS replays do not replace hosted validation or establish an external cause. Candidate tree remains 9bad791482638be26d3ab306653fd279791f449b.

- `CURATOR_CONFORMANCE_ROOT=<CI-pin>/conformance/v1 go test ./internal/install -run '^TestAuthoritativeDryRunCasesMutateNothingPersistent$/^compiled-cache-miss-is-read-only$' -count=1 -v -timeout=4m`: exit 1. The named subcase passes, but the family completeness checker rejects the filtered run. Not counted as passing evidence.
- `CURATOR_CONFORMANCE_ROOT=<CI-pin>/conformance/v1 go test ./internal/install -run '^TestAuthoritativeDryRunCasesMutateNothingPersistent$' -count=1 -v -timeout=4m`: exit 0; all 3 locally applicable cases pass.
- `CURATOR_CONFORMANCE_ROOT=<CI-pin>/conformance/v1 go test ./cmd/curator -run '^TestReviewProbeLegacyLaneIdentityInvalid$' -count=1 -v -timeout=5m`: exit 0.

The two original failures did not reproduce in these local whole-test runs. The second diagnostic (`toolchain_mutated`) wraps any revalidation error in internal/godriver/session.go; it is not proof that the compiler changed. Underlying hosted cause remains unknown.

`gh run rerun 39004671814 --job 117314584571` initially refused while the run was active. The same command succeeded after the run completed. Only that failed job is retried; successful first-attempt jobs are retained. No new snapshot or source edit.

## diagnostic-dryrun.log

```text
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only
=== NAME  TestAuthoritativeDryRunCasesMutateNothingPersistent
    coverage.go:234: published-case coverage: family "install/dry-run-cases" produced a result for unpublished case ""
    coverage.go:236: published cases install/dry-run-cases: 0 driven, 0 known-gap, 0 bound, 0 skipped, 0 total
--- FAIL: TestAuthoritativeDryRunCasesMutateNothingPersistent (10.39s)
    --- PASS: TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only (10.38s)
FAIL
FAIL	github.com/relux-works/curator/internal/install	10.748s
FAIL

```

## diagnostic-dryrun-family.log

```text
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent/project-upgrade
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent/global-upgrade
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only
=== NAME  TestAuthoritativeDryRunCasesMutateNothingPersistent
    coverage.go:236: published cases install/dry-run-cases: 3 driven, 0 known-gap, 0 bound, 0 skipped, 3 total
--- PASS: TestAuthoritativeDryRunCasesMutateNothingPersistent (8.65s)
    --- PASS: TestAuthoritativeDryRunCasesMutateNothingPersistent/project-upgrade (0.39s)
    --- PASS: TestAuthoritativeDryRunCasesMutateNothingPersistent/global-upgrade (0.64s)
    --- PASS: TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only (7.62s)
PASS
ok  	github.com/relux-works/curator/internal/install	8.873s

```

## diagnostic-legacy.log

```text
=== RUN   TestReviewProbeLegacyLaneIdentityInvalid
--- PASS: TestReviewProbeLegacyLaneIdentityInvalid (8.31s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	8.674s

```
