# Hosted first-attempt failures

Candidate tree: 9bad791482638be26d3ab306653fd279791f449b. Snapshot: be87653a94f23f39acbaa3e965a8f424cf769b60.

Job: https://github.com/relux-works/curator/actions/runs/39004671814/job/117314584571

Go test exit 1; platform-case gate exit 0. This is a failed run, not green evidence. Root cause beyond the reported observations is unknown. No external-outage exception is asserted.

## ('github.com/relux-works/curator/internal/install', 'TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only')

```text
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only
    dryrun_conformance_test.go:297: project 0 reported build outcome "toolchain-unavailable" (reason "go-v1 process_timeout: Go probe exceeded its deadline", diagnostic "process_timeout"), which the published case does not admit: [cache-hit would-preflight-and-build would-rebuild-untrusted-cache corrupt unsupported]
--- FAIL: TestAuthoritativeDryRunCasesMutateNothingPersistent/compiled-cache-miss-is-read-only (167.07s)
```

## ('github.com/relux-works/curator/internal/install', 'TestAuthoritativeDryRunCasesMutateNothingPersistent')

```text
=== RUN   TestAuthoritativeDryRunCasesMutateNothingPersistent
    coverage.go:234: published-case coverage: failing published case install/dry-run-cases/compiled-cache-miss-is-read-only is not listed in the gap ledger: case test failed
    coverage.go:236: published cases install/dry-run-cases: 2 driven, 0 known-gap, 0 bound, 0 skipped, 2 total
--- FAIL: TestAuthoritativeDryRunCasesMutateNothingPersistent (171.15s)
```

## ('github.com/relux-works/curator/cmd/curator', 'TestReviewProbeLegacyLaneIdentityInvalid')

```text
=== RUN   TestReviewProbeLegacyLaneIdentityInvalid
    draft_transport_test.go:814: [status app]: stderr misses the frozen diagnostic "error: skill-a.ssh-cmd: build_repository_source_unavailable: exact external source is unavailable":
        warning: security_posture_permissive: security_posture is permissive; set security_posture: hardened in the machine configuration to adopt the hardened defaults
        error: go-v1 toolchain_mutated: cannot revalidate toolchain
--- FAIL: TestReviewProbeLegacyLaneIdentityInvalid (245.60s)
```
