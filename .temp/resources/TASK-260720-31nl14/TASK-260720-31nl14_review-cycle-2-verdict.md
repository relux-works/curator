# TASK-260720-31nl14 review cycle 2 verdict

## Verdict

Changes requested. Route to to-dev.

## Verified cycle-1 fixes

Directory desired digests now bind root permissions, and rollback plus cleanup reject root-metadata drift while retaining recovery state. Canonical live, desired, backup, and rollback paths reject nesting, symlink aliases, hard-link aliases, direct sidecar collisions, and native-Windows case aliases.

## Findings

1. P1 - Cleanup tomb and journal namespaces remain unreserved, and removeDurably deletes an existing tomb without provenance validation. namespace.go:18-39 validates only live, staged, backup, and rollback paths. journal.go:312-339 derives path + .delete later and removes any existing tomb at lines 320-326. engine.go:640-650 performs all target digest checks before cleanup, then removes sidecars and finally the journal. A valid-looking two-target plan can set target 1 live to target 0 BackupPath + .delete; both desired digests pass the cleanup precheck, then deleting target 0 backup first removes target 1 live as a supposed crash tomb, and commit returns success with target 1 missing. A target at journalPath(transactionID) + .delete is likewise deleted by final journal cleanup. This violates independent target boundaries, unknown-state preservation, and successful-commit semantics. Reject target live/sidecar/tomb paths overlapping every target cleanup tomb and the engine journal namespace before the first journal write. removeDurably must treat simultaneous original plus tomb or a tomb whose expected identity/digest cannot be proven as implementation-corruption and preserve current bytes. Add commit-cleanup and restart-cleanup regressions proving no target or unknown tomb is deleted.

2. P2 - Case-alias rejection is Windows-only even on a case-insensitive Darwin filesystem. namespace.go:134-138 uses EqualFold only when runtime.GOOS is windows, and validation_windows_test.go:1-23 excludes the regression from Darwin. The review probe in the actual workspace reports filesystem_case_sensitive=false. Two absent live paths named Target and target therefore produce distinct canonical keys, both Stat calls return not-exist, and Prepare accepts aliases that cannot be independent; commit later fails only after mutating the first target and entering rollback. Detect case behavior for the containing filesystem or conservatively reject case-fold aliases on supported case-insensitive platforms, and add a native-Darwin regression that is conditional on filesystem behavior.

## Validation evidence

Pass: go test ./internal/transaction -count=1; go test -race ./internal/transaction -count=1; go vet ./internal/transaction; make check; go test -race ./... -count=1; make build on Darwin arm64; complete Linux amd64 and Windows amd64 compile graphs with go test -exec=true; golangci-lint v2.4.0 scoped to internal/transaction with 0 issues; gofmt, git diff --check, and staged-file checks. task-board validate reports the same 13 inherited board-wide issues and none belongs to this task. Native Windows runtime remains unavailable; compilation is not runtime evidence and TASK-260720-1zl1cj remains unchanged. No product code was modified during review.