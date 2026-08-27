# TASK-260720-1zl1cj review verdict cycle 3

## Verdict

Changes requested; route to implementation rework (`to-dev`).

## Correctness finding

The cycle-3 Windows canonicalization still does not preserve physical manager-home identity on per-directory case-sensitive filesystems.

`internal/managerlock/identity_windows.go:15-33` queries `FileCaseSensitiveInfo` only for the complete existing path (or longest existing prefix), then applies that one leaf flag to the entire pathname: a false flag uppercases every component and a true flag preserves every component. Windows case sensitivity is per directory, so the comparison rule for a component is governed by its parent directory, not by the leaf directory's flag.

Concrete failure shape: a case-sensitive parent can contain two distinct child directories named `Foo` and `foo`, while those child directories themselves are case-insensitive (Windows applications can create non-case-sensitive children in a case-sensitive tree, and an empty child flag can also be changed). `New(parent\\Foo)` and `New(parent\\foo)` both observe a false leaf flag and uppercase the complete paths to the same identity. The managers then share `processState` and derive the same lock root despite referring to distinct physical homes; the uppercased canonical path can also address a third spelling under the case-sensitive parent. This violates canonical physical identity and the explicit cycle-2 requirement not to break case-sensitive filesystems.

The same one-flag model is insufficient for a multi-component nonexistent suffix because each newly created intermediate directory can have different case semantics. The new regression only covers an ordinary case-insensitive temporary root and explicitly skips a case-sensitive root, so it cannot detect either failure.

Microsoft documents both that the flag is per directory and that Windows-created directories in case-sensitive trees may be non-case-sensitive: https://learn.microsoft.com/en-us/windows/wsl/case-sensitivity

## Required rework

- Canonicalize Windows paths component by component according to the containing directory's lookup semantics, or use an equivalent stable physical-identity design. Do not apply the leaf directory's flag to the whole path.
- Add a native Windows regression with a case-sensitive parent and two distinct case-only child homes whose own flags are case-insensitive. Prove distinct `Home`, lock roots, and process-order state, and prove no lock state is redirected to an uppercase third spelling.
- Add a case-sensitive-prefix/multi-component first-use regression that observes the actual created intermediate-directory semantics and proves aliases contend while distinct physical homes remain distinct.
- Preserve the ordinary case-insensitive first-use alias regression and the Unix symlink-prefix regression.
- Run the subprocess suite on a native Windows runner; cross-compilation alone does not execute these paths.

## Passing evidence

- `go test -race -cover ./internal/managerlock -count=1 -v`: passed on Darwin/arm64, 82.4% statement coverage, including Unix subprocess contention, cancellation, and abnormal exit.
- `go test -race ./... -count=1`: passed repository-wide.
- `make check`: passed (`go vet ./...`, `go test ./...`, `gofmt` check).
- `make build`: passed; native Darwin/arm64 CLI produced.
- Linux/amd64 and Windows/amd64 managerlock test binaries compiled; the Windows binary contains `TestWindowsMissingHomeCaseAliasesShareIdentityAndContention`.
- `go vet ./internal/managerlock`, focused `gofmt`, `git diff --check`, and no-staging checks passed.
- Native Windows execution remains unavailable on this Darwin host; no Windows runtime pass is claimed.

No product code was modified during review.
