# TASK-260720-1zl1cj review rework cycle 4

## Review defect addressed
Windows canonicalization no longer applies the leaf directory case-sensitivity flag to the whole path. Existing components are normalized one at a time using the lookup semantics of their containing directory. Real acquisitions create the configured home and then re-resolve its physical identity before selecting process state or any lock path, which handles multi-component first use even when created-directory semantics differ. Dry-run returns before this preparation and creates no state. Manager identity fields are synchronized for concurrent first use.

## Regression coverage
Added Windows-only coverage for a sensitive parent containing distinct Foo and foo homes whose own flags are insensitive; the test proves distinct Home values, lock roots, process state, independent home locks, and no redirected uppercase sibling. Added a sensitive-prefix multi-component first-use test that inspects each created directory flag, derives a real semantic alias, proves subprocess contention, stabilizes a manager constructed before creation, and keeps a distinct first component independent. Added a portable concurrent first-use race regression. Existing ordinary Windows case-alias and Unix symlink-prefix regressions remain.

## Validation
PASS: go test -race -cover ./internal/managerlock -count=1, coverage 82.8 percent. PASS: go test -race ./... -count=1. PASS: make check. PASS: make build. PASS: native go vet, focused go vet, GOOS=windows go vet, gofmt, git diff --check, and no staged changes. PASS: Linux amd64 and Windows amd64 managerlock test compilation. PASS: Linux amd64 and Windows amd64 curator build. go tool nm confirms all three Windows regressions and TestManagerLockHelper are present in the Windows test executable.

## Platform boundary
Native Windows runtime execution was not run: the host is Darwin arm64 and has no Wine, Parallels, VirtualBox, Lima Windows runtime, or repository self-hosted Windows runner. Existing GitHub-hosted Windows CI can only exercise landed or pushed code; this task explicitly forbids staging, committing, or pushing. The compiled Windows test executable is attached for native execution. golangci-lint is not installed; task-defined go vet and gofmt lint gates pass.

## Provenance
Worktree HEAD and origin/main are 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8. No files are staged. Windows test executable SHA-256: e76fb84cf807bcbfb671ad1fd51e4c90837b9d56e237f8f74bf6490f74b43e92.