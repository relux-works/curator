# BUG-260924-5p8b0z review verdict — CR rev1: ACCEPTED

Reviewer: claude-opus-5-5 (low). Disposable clone /tmp; `git write-tree` = 4fd054685f5ef04d010cc47880e09b563dd5efe7 (the exact candidate).

## Findings
1. Fix `internal/scriptworker/exec.go:79`: the System32 hard-link trust root is set only when all of these hold: platform windows, the default search list is in use, and `managerEnvironment != nil` (an explicit manager-captured snapshot). A nil environment still uses the ambient SYSTEMROOT for search, but that value can no longer grant hard-link trust. In production the snapshot is always non-nil: `cmd/curator/main.go:203 Environ: os.Environ()` → `launcher.go:226` → `client.go:87 HostEnv`. So the one accepting case still accepts in production.
2. The vector matches the spec: the testdata equals curator-spec dcc7f015 `conformance/v1/vectors/script-host-execution-policy.json` `.executable_identity_cases` (python equality True, 8 cases). The file's sha256 `124e0075…92a2` is pinned in the test.
3. Local run on darwin with the GOOS seam (`deriveProfileForPlatform(…, "windows")`, the path Launch uses): all 8 subtests PASS. The full `internal/scriptworker` package passes (31.4 s) and `go vet` is clean.
4. Mutant (drop `&& managerEnvironment != nil`), which I re-applied myself: it FAILS `windows-exec-uncaptured-systemroot-hardlinks` with "accepted = true, want false", so it is killed.
5. Hosted gate run 35987255178 (from the validation log, exit 0): the windows-latest `test/go-test.json` contains a `pass` row for the parent test and for all 8 subtests.
6. No CHANGELOG edit and no stray files. Changed paths: exec.go, the test, and the testdata.

## Residuals (stated bounds, not blocking)
- R1: the "uncaptured" case is modelled as a nil HostEnv plus an ambient SYSTEMROOT. Production has no other source of a non-manager SYSTEMROOT that reaches this decision, because env_read passthrough does not feed managerEnvironment.
- R2: I did not re-run the mutant on a Windows host. The decision depends only on the platform string, so the darwin kill covers the same code.
- R3: the comment at `capabilities.go:124-125` ("Nil … is what production passes") is stale. Production passes os.Environ(). Only the doc is wrong.
- R4: `windows-exec-noncomponent-store-hardlinks` and `windows-exec-unowned-file-hardlinks` are still inverted in the test (the want=true override at test:88-91), as owned R5 gaps. This is outside this bug's scope.
- Gap-ledger: no ledger row for this case exists on the base; the in-test override list excludes it. That satisfies the AC.

Logbook DoD item: this resource records the findings.
