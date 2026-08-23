# TASK-260720-1zl1cj review verdict — cycle 5

## Verdict

Changes requested. Route to `to-dev` for test rework and another native Windows reviewer cycle.

## Blocking review finding

The exact current candidate's native Windows package suite exits with status 1. Two nominally portable tests encode a Unix-style casing assumption that contradicts the Windows canonicalization implemented by the package:

- `internal/managerlock/managerlock_test.go:132-135` constructs `want` from `filepath.EvalSymlinks` plus the original lowercase missing suffix. On a case-insensitive Windows path, `New` intentionally returns the normalized uppercase identity, so `TestMissingHomeUsesCanonicalExistingPrefix` fails before it can check pre/post-creation stability.
- `internal/managerlock/managerlock_test.go:164-170` repeats the same assumption for an aliased existing prefix, so `TestMissingHomeBelowSymlinkKeepsIdentityAndContends` fails before its subprocess contention assertion.

Observed failures on `DESKTOP-3PBO632`, Windows 10.0.19045, PowerShell 5.1:

```text
--- FAIL: TestMissingHomeUsesCanonicalExistingPrefix
missing home identity = "C:\\USERS\\...\\MISSING\\CURATOR-HOME", want ... "C:\\Users\\...\\missing\\curator-home"
--- FAIL: TestMissingHomeBelowSymlinkKeepsIdentityAndContends
missing aliased home identity = "C:\\USERS\\...\\REAL\\MISSING\\CURATOR-HOME", want ... "C:\\Users\\...\\real\\missing\\curator-home"
FAIL
```

This is ordinary implementation/test rework, not an external stop-the-line blocker. The acceptance criteria and task DoD require Windows tests green.

## Required rework

1. Make both portable tests assert the package's platform canonical identity semantics rather than raw `EvalSymlinks` spelling. Preserve their substantive checks: stable identity before/after creation and contention through an aliased existing prefix.
2. Recompile from the unchanged candidate provenance and run the complete `internal/managerlock` test binary on native Windows; attach a zero-exit verbose log and exact SHA-256.
3. Route the task through another reviewer cycle. Do not weaken or remove the Windows implementation behavior to satisfy casing-only expectations.

## Additional Windows evidence

- Fresh cross-compile SHA-256: `e76fb84cf807bcbfb671ad1fd51e4c90837b9d56e237f8f74bf6490f74b43e92`; byte-identical to the previously attached candidate binary.
- `TestWindowsMissingHomeCaseAliasesShareIdentityAndContention` passed natively.
- Project/home/build-key subprocess contention, cancellation, and abnormal-child-exit tests passed natively.
- `TestWindowsMixedCaseSensitivityKeepsDistinctHomes` and `TestWindowsCaseSensitivePrefixMultiComponentFirstUse` skipped because this runner returned `The request is not supported` when enabling per-directory case sensitivity. This remains an explicit coverage caveat for a capable NTFS/Windows environment; it did not cause the suite's nonzero exit.
- Remote task directory and binary were removed after execution and absence was verified.

## Independent local validation

- `go test -race -cover ./internal/managerlock -count=1 -v`: pass, 82.8% statement coverage.
- `go test -race ./... -count=1`: pass.
- `make check`: pass.
- `go build ./...`: pass.
- `go vet ./...`: pass.
- Linux amd64 and Windows amd64 managerlock test compilation: pass.
- `gofmt -d internal/managerlock`: clean.
- `git diff --check`: clean.
- `git diff --cached --quiet`: pass; no staged changes.

Candidate base remains detached `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` with the task handoff's imported accepted product diff plus `internal/managerlock`; review made no product-code changes.
