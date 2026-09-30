# TASK-260930-1kylpg results — named-path absence pinned at the boundary

Tests only; no production code changed (`git diff` empty; two new untracked test files).

## Rows
1. `internal/pathboundary/named_absence_test.go`
   - `TestNamedRoutesRejectMissingTarget`: ValidateWithinWithOwner / ValidateRouteWithOwner / ValidateLeafWithOwner on a missing named target (`root/missing`, `root/store/missing`, missing intermediate `root/missing/entry`) → `*Failure{Check: regular_types}` wrapping `fs.ErrNotExist`. Platform-neutral (no unix-only helpers); Windows compile checked with `GOOS=windows go vet` (rc=0); not executed on Windows locally — hosted gate is the arbiter.
   - `TestValidateTreeRejectsMissingWalkTarget`: validateTree on a missing walk target → regular_types Failure (pins the `path == target ||` guard in vanished() directly).
2. `internal/envprofile/named_absence_boundary_test.go` — `TestResolveRefusesMissingLockNamedStoreEntryAtBoundary`: provisioned fixture, remove the lock-named context store entry, then
   - boundary step `validateNamedStoreBoundaries` returns failure on that entry path, check `regular_types`, wrapping ErrNotExist;
   - production entry `Resolve` returns `environment_store_untrusted ... failed regular_types check`, and NOT `failed pin_hash check`; no fragment; manager state hash unchanged.
   Distinguishability: yes — the boundary step reports `regular_types`, pin recomputation reports `pin_hash`; the check name in the diagnostic is observable.

## Mutant (three named-route Lstat loops `if IsAbsent(err) { return nil }` + drop `path == target ||` in vanished())
- Before (new rows hidden): `go test ./pathboundary` ok; `go test ./envprofile -timeout 45m` rc=0 (807 s) → SURVIVES. (First attempt hit Go's default 10 m timeout under host load; no test failures; rerun with longer timeout.)
- After (rows restored, mutant in place): rc=1 — all 9 route cases, validateTree row, and envprofile boundary-step row fail → KILLED.
- Mutant reverted by copying back the saved original.

## Validation on unmodified production code
- `go test ./pathboundary -count=1` rc=0
- `go test ./envprofile -run 'MissingLockNamed|StoreBoundary|ResolveRepairKeeps' -count=1` rc=0 (full envprofile suite green rc=0 in the before-run; production code identical)
- `gofmt -l` clean; `go vet` rc=0; `GOOS=windows go vet` rc=0

## Gate fix 1 (Windows DACL fixture)
- Windows run 36726325163 failed because t.TempDir() root is not owner-only; enclosing-root DACL check fired first.
- Fix (fixture only): pathboundary named-absence tests now create root via privatedir.MakeAll (privateRoot helper). No skip, assertion unchanged. Envprofile row already used privatedir-backed fixtures.
- `GOOS=windows go vet ./internal/pathboundary ./internal/envprofile` exit 0
- `go test ./internal/pathboundary -run 'NamedRoutes|MissingWalkTarget' -count=1` exit 0
- `go test ./internal/envprofile -run 'NamedAbsence|LockNamed' -count=1 -v` → TestResolveRefusesMissingLockNamedStoreEntryAtBoundary PASS, ok
- Mutant evidence from rev1 retained unchanged (fixture change does not affect it on darwin). Windows confirmation pends hosted gate.
