# BUG-260923-krcm6m review verdict — CR rev3: ACCEPTED

Delta 1511b345..bbeb8cda: CHANGELOG.md and internal/registry/registry_test.go (test-only). No product change.

1. Root cause. The hosted artefact (run 35901867517, `_hosted_windows_evidence.json`) shows `registry_test.go:734 skew 0s offset 1s: refused=false` with the subtest taking 1.55 s. The test passes a fixed `now` and an instant fetch. The hidden time source is `time.Since(start)` in internal/registry/snapshot.go:178, and `start` is set at function entry (snapshot.go:89). That is before `loadSnapshotStateCatalog` creates and fsyncs the empty catalog in a fresh temp dir (snapshot.go:300/407/453/475). So on Windows, first-use fsync time of more than 1 s is added to the allowance, and `now+1s` gets through. The 1.55 s duration fits this cause. It is an inference from the duration, not a timing breakdown; I accept it. Each row gets its own temp dir, which rules out cache reuse.
2. The fix is at the cause. It seeds the empty catalog before the call (writeSnapshotStateCatalog), so catalog I/O is no longer inside the elapsed window. The bound is not widened, nothing is retried, and the production threshold is unchanged. The test still calls `CheckSnapshotsWithPolicy` (persist lane), and the inside, exact, +1 s and far-future rows keep their assertions.
3. Narrowing mutant, re-applied by me in a disposable clone: `now.Add(clockSkew).Add(time.Second).Add(time.Since(start))` → FAIL `registry_test.go:744: skew 5m0s offset 1s: refused=false want true`. The mutant is killed. I restored the file afterwards.
4. Repetition, run by me: `go test ./internal/registry -run TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew -count=50` → ok (129 s, darwin).
5. Rev3 validation log: remote gate run 35935995625 passed, including Test (windows-latest), Race, Lint and Gate self-test.
6. CHANGELOG has a Fixed entry.

Residual (not blocking): in production, the latency allowance also covers first-use catalog creation. This is intended per the doc comment at snapshot.go:62-70.
