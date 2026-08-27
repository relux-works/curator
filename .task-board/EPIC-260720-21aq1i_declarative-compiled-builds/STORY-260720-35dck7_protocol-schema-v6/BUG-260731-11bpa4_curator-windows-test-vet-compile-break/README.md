# BUG-260731-11bpa4: curator-windows-test-vet-compile-break

## Description
Curator Test (windows-latest) fails before any test runs: go vet rejects the package with vet.exe: internal\runtimestore\targets_windows_test.go:97:14: undefined: decodeHelperOutput. The windows-only test file references a helper that does not exist in the windows build, so the package does not compile there. Pre-existing on main cfffd7cd and unrelated to any protocol vector; it was masked until BUG-260731-3gm8kc PR 9 repaired the toolchain-identity gate that had been failing every Go job at step 4. Evidence: run 30615765014 job 91108467247, plus the isolated control run on branch ci/goenv-control-BUG-260731-3gm8kc which carries only the toolchain-identity repair.

## Scope
internal/runtimestore windows-only test helpers. Restore the missing helper or the call site so the windows build compiles; do not delete the case to make the gate pass.

## Acceptance Criteria
go vet and go test pass for internal/runtimestore on windows-latest in Curator CI.
