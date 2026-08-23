# TASK-260720-31nl14 review cycle 4 rework

## Resolution

The cleanup journal now durably records a canonical sorted per-entry manifest before renaming any owned sidecar to its deletion tomb. Recovery validates the complete remaining tomb against recorded file digests, file and directory modes, kinds, paths, and parent structure before deletion. Missing recorded entries are treated as already durably removed; unrecorded, replaced, unsafe, or metadata-mismatched entries return ErrImplementationCorruption without changing current bytes. The prior unrecorded owned-tomb bypass was removed.

## Regression coverage

Direct and restarted recovery now cover commit cleanup and rollback cleanup at PointAfterCleanupRename and PointDuringCleanupRemoval, for both an added foreign file and a replaced recorded file. Every corruption branch preserves the foreign bytes, consumer target, journal, and recovery state; restoring the recorded tomb state allows deterministic recovery to finish. The reviewer exported-API probe now reports implementation-corruption and foreign_exists_after_recovery=true. Existing unmodified full and partial tomb recovery continues to pass.

## Verification

Passed on native Darwin arm64: focused transaction tests; focused race; focused vet; 77.3 percent statement coverage; reviewer probe; subprocess crash recovery; make check across the complete imported graph; full repository race; native make build and curator version runtime; Linux amd64 and Windows amd64 complete compile graphs via go test -exec=true; golangci-lint v2.4.0 with 0 issues; repeated new regression tests; gofmt; git diff --check; and no staged files. Repository-wide commands used the established task-local go.test.mod because the exact-base worktree intentionally leaves the tracked tuitestkit submodule unmaterialized. An initial unqualified full run failed only at package loading for that absent replacement and was rerun successfully with the established modfile.

Native Windows runtime execution remains unavailable on this Darwin host. Windows compilation is not runtime evidence, and the inherited TASK-260720-1zl1cj qualification gate remains preserved. task-board validate still reports the same 13 inherited unrelated issues: 12 legacy EPIC-260712 dependency links and one orphan TASK-260713-7a9c1e resource.