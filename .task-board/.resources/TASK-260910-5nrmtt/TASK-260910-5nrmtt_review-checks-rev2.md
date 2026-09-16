# Independent revision 2 checks

Baseline exit 0; repro and each mutant exit 1. All overlays compiled from candidate bytes; repository files never modified.

## baseline

```text
ok  	github.com/relux-works/curator/internal/buildrepo	50.314s
```

## repro

```text
--- FAIL: TestReviewRev2MixedUnknownMustNotFallback (4.11s)
    --- FAIL: TestReviewRev2MixedUnknownMustNotFallback/fatal:_unable_to_access_'https://fixture.test/repository.git/':_Could_not_resolve_host:_fixture.test_fatal:_unexpected_protocol_response (1.14s)
        transport_test.go:1737: fail-closed diagnostic produced 2 fetches, err=<nil>
    --- FAIL: TestReviewRev2MixedUnknownMustNotFallback/fatal:_unable_to_read_object_0123456789012345678901234567890123456789_fatal:_connection_timed_out (1.13s)
        transport_test.go:1737: fail-closed diagnostic produced 2 fetches, err=<nil>
    --- FAIL: TestReviewRev2MixedUnknownMustNotFallback/fatal:_cannot_create_temporary_file_'connection_timed_out':_Permission_denied (1.20s)
        transport_test.go:1737: fail-closed diagnostic produced 2 fetches, err=<nil>
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	5.064s
FAIL
```

## m1

```text
--- FAIL: TestClassifyFetchOutput (0.00s)
    --- FAIL: TestClassifyFetchOutput/audit-canary (0.00s)
        transport_test.go:121: ClassifyFetchOutput = "unknown", want "audit"
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	1.104s
FAIL
```

## m1b

```text
--- FAIL: TestResolvedTransportAmbiguousFailureMustNotFallbackWhenAlternateReady (1.83s)
    --- FAIL: TestResolvedTransportAmbiguousFailureMustNotFallbackWhenAlternateReady/audit-timeout (0.72s)
        transport_test.go:1400: ambiguous failure fell back to a ready alternate and succeeded
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	2.246s
FAIL
```

## m2

```text
--- FAIL: TestResolvedTransportTruncatedStderrFailsClosed (1.08s)
    transport_test.go:1471: truncated evidence fell back to a ready alternate and succeeded
FAIL
FAIL	github.com/relux-works/curator/internal/buildrepo	1.672s
FAIL
```
