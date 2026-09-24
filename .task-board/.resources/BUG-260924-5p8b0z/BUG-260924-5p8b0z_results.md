# BUG-260924-5p8b0z — results

## Change

The Windows System32 hard-link allowance in `internal/scriptworker/exec.go` now requires an explicit manager environment snapshot as well as the default Windows search list. A nil environment may still provide ambient search directories, but it no longer grants captured-SystemRoot trust.

Added the exact eight `executable_identity_cases` from curator-spec commit `dcc7f015` as an embedded fixture. Its SHA-256 is `sha256:124e00757b3add8c2ba639a8403cba97eec7f5d1bdd923d14b51db9a104292a2`. `TestExecutableIdentityCasesAtProductionEntry` drives declared exec cases through `deriveProfileForPlatform` and interpreter cases through `ResolveInterpreter`, using the existing injected-Windows/environment seam. The uncaptured-SystemRoot case is rejected; the two neighboring noncomponent-store and unowned-file gaps retain their prior observed outcomes.

## Gap accounting

The assigned Story base has no `.github/ci/conformance-gaps.tsv`. The attached TASK-260922-18ex37 rev3 candidate ledger contains only the two neighboring rows for noncomponent-store links and unowned files; it has no uncaptured-SystemRoot row. This change adds no gap for the target case, and its test has no target-case exception: it must match the vector's `accepted: false` result. The neighboring cases remain unchanged.

## Validation

| Command | Exit | Result |
| --- | ---: | --- |
| Fixture comparison with `dcc7f015:conformance/v1/vectors/script-host-execution-policy.json` | 0 | Exact eight-case array match |
| `go test ./internal/scriptworker -run '^TestExecutableIdentityCasesAtProductionEntry$' -count=1` | 0 | All eight cases pass through production resolvers |
| Same targeted test with the `managerEnvironment != nil` condition removed | 1 | Expected red: uncaptured-SystemRoot case accepted=true, wanted=false; mutant killed |
| `go test ./internal/scriptworker` after restoring the condition | 0 | Package suite passes |
| `golangci-lint run ./internal/scriptworker` | 0 | 0 issues |
| `go vet ./internal/scriptworker` | 0 | Clean |
| `gofmt -d internal/scriptworker/exec.go internal/scriptworker/exec_identity_conformance_test.go` | 0 | No output |
| `git diff --check` | 0 | Clean |
| `go build -o $TMPDIR/BUG-260924-5p8b0z-evidence/curator ./cmd/curator` | 0 | Build succeeds |
| `GOOS=windows GOARCH=amd64 go test -c -o $TMPDIR/BUG-260924-5p8b0z-evidence/scriptworker.test.exe ./internal/scriptworker` | 0 | Windows test binary compiles |

The mutant was exercised locally on macOS through the Windows platform seam. A native hosted Windows runtime lane was not run locally; the configured hosted landing gate runs at handoff.

## Worktree evidence

The configured integration branch is `main`. Fresh `git ls-remote` advertised `5b326aa384d9e3f3a7c047e7b228107e84c90df9`; the exact-ref fetch and local `HEAD`, `main`, and `origin/main` all matched that OID. The Story worktree record reported base `main`, tip `5b326aa384d9e3f3a7c047e7b228107e84c90df9`, and a clean tree before changes. Current product changes are limited to `exec.go`, the new production-entry test, and the pinned fixture. No CHANGELOG or LOGBOOK file was edited.

## CHANGELOG entry (for release prep)

Narrow the Windows System32 executable hard-link allowance to resolutions backed by an explicitly captured manager SYSTEMROOT; ambient values do not grant the exception.
