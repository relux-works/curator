# TASK-260917-16l2md — hosted gate failure on Change Request revision 2 (run 35276210791)

Extracted by the orchestrator from the CI evidence. The vector suites stay
green at `SPEC_PIN dced9b8` on every lane; the rev-1 `TestEnvStatusMatrix`
failure is fixed. Two new failures, both in the rev-2 identity tests for
the E4 trust roots:

## 1. windows-latest — `cmd/curator TestUmbrellaTrustRootShortSpellingIdentity`

```
umbrella_trustroot_windows_test.go:70: resolved "C:\Users\runneradmin\...\001\curator-run.exe",
                                       want     "C:\Users\RUNNER~1\...\001\curator-run.exe"
```

The production code now canonicalizes the 8.3 short spelling to the long
one (correct); the test still expects the raw short spelling back. Fix the
expectation: the test must assert IDENTITY, not spelling — compare the
resolved provider against the expected path after canonicalizing both (or
`os.SameFile`), and assert the posture row is current/trusted. Never make
the code return the short spelling to satisfy the test.

## 2. ubuntu-latest (Test and Race) — platform-case gate, not `go test`

```
FAIL  skip with an unrecognised reason on linux: cmd/curator :: TestUmbrellaTrustRootCaseVariantIdentity
      reason: volume is case-sensitive: stat /tmp/.../caseroot: no such file or directory
      add it to .github/ci/skip-classes.tsv with a class, or fix the case.
```

`go test` is green; the new case-variant identity test skips on a
case-sensitive volume with a free-text reason. Every skip on this repository
must belong to a class in `.github/ci/skip-classes.tsv` and carry a ledger
row in `.github/ci/platform-cases.tsv` (host-capability class, naming the
absent capability — a case-insensitive filesystem — never the GOOS). Do
that: register the reason under the appropriate existing class (or add a
class row the way the file documents) plus the ledger row, and keep the
test executing on Windows/macOS where the volume is case-insensitive.
Alternatively make the test create a case-insensitive fixture where
possible — but the skip must be classed either way.

Re-run the narrow gates with the rc.12 root and hand off again; the runtime
re-runs the hosted gate.
