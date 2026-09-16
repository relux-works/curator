## attack.log (exit 1)

```text
--- FAIL: TestReviewResolvedMustRetainAdmission (4.57s)
    --- FAIL: TestReviewResolvedMustRetainAdmission/missing-ssh-wrapper (1.43s)
        transport_test.go:1256: resolved accepted missing-ssh-wrapper; legacy refused build_repository_identity_invalid: SSH requires the exact manager wrapper; fetches=1
    --- FAIL: TestReviewResolvedMustRetainAdmission/missing-ssh-credentials (1.25s)
        transport_test.go:1256: resolved accepted missing-ssh-credentials; legacy refused build_repository_ssh_credential_missing: SSH build repositories require an operator identity or agent; fetches=1
    --- FAIL: TestReviewResolvedMustRetainAdmission/missing-https-broker (0.57s)
        transport_test.go:1256: resolved accepted missing-https-broker; legacy refused build_repository_identity_invalid: HTTPS requires a manager credential broker; fetches=1
    --- FAIL: TestReviewResolvedMustRetainAdmission/invalid-ref-kind (0.60s)
        transport_test.go:1256: resolved accepted invalid-ref-kind; legacy refused build_repository_identity_invalid: invalid substitution ref kind; fetches=1
--- FAIL: TestReviewAmbiguousFailureMustNotFallback (2.41s)
    --- FAIL: TestReviewAmbiguousFailureMustNotFallback/fatal:_cannot_create_temporary_file:_Permission_denied (0.60s)
        transport_test.go:1268: ambiguous/local/integrity error caused 2 fetches (err=<nil>)
    --- FAIL: TestReviewAmbiguousFailureMustNotFallback/remote:_audit_denied:_operation_timed_out (0.73s)
        transport_test.go:1268: ambiguous/local/integrity error caused 2 fetches (err=<nil>)
    --- FAIL: TestReviewAmbiguousFailureMustNotFallback/error:_object_hash_mismatch_fatal:_connection_timed_out (0.76s)
        transport_test.go:1268: ambiguous/local/integrity error caused 2 fetches (err=<nil>)
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	7.694s
FAIL
```

## deadline.log (exit 1)

```text
--- FAIL: TestReviewDeadlineIncludesChildPipes (6.53s)
    transport_test.go:1280: 2s total deadline returned after 5.730310214s: build_repository_source_unavailable: exact source fetch failed
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	7.131s
FAIL
```

## m1.log (exit 1)

```text
--- FAIL: TestSemanticFallbackGateDecidesAllPublishedCases (0.00s)
    --- FAIL: TestSemanticFallbackGateDecidesAllPublishedCases/fallback-unknown (0.00s)
        transport_test.go:50: gate("unknown") advance = true, want false
--- FAIL: TestResolvedTransportFailClosedClassesStopAfterOneFetch (11.21s)
    --- FAIL: TestResolvedTransportFailClosedClassesStopAfterOneFetch/unknown (2.30s)
        transport_test.go:765: fail-closed failure succeeded
    --- FAIL: TestResolvedTransportFailClosedClassesStopAfterOneFetch/empty (1.81s)
        transport_test.go:765: fail-closed failure succeeded
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	11.852s
FAIL
```

## m2.log (exit 1)

```text
--- FAIL: TestClassifyFetchOutput (0.00s)
    --- FAIL: TestClassifyFetchOutput/host-key (0.00s)
        transport_test.go:111: ClassifyFetchOutput = "unknown", want "host-key"
    --- FAIL: TestClassifyFetchOutput/host-key-beats-auth (0.00s)
        transport_test.go:111: ClassifyFetchOutput = "auth", want "host-key"
--- FAIL: TestResolvedTransportFailClosedClassesStopAfterOneFetch (11.38s)
    --- FAIL: TestResolvedTransportFailClosedClassesStopAfterOneFetch/host-key (1.12s)
        transport_test.go:775: records = [{Index:1 URL:https://fixture.test/repository.git Transport:https Provider: Class:unknown NetworkAttempted:true Succeeded:false}], want one host-key record
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	12.547s
FAIL
```
