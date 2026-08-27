# TASK-260720-31nl14 review-cycle 1 rework validation

## Outcome

The two review-cycle P1 findings are addressed in `internal/transaction`:

- Directory target digests bind root-directory permission metadata, and rollback/cleanup preserve a live target whose root metadata no longer matches the recorded desired digest.
- Plan and journal validation reject overlapping live/staged/backup/rollback namespaces, symlink and hard-link aliases, noncanonical sidecars, and native-Windows case aliases before journal creation.

The implementation remains scoped to the transaction package on top of the imported manager-lock candidate. It does not refactor `install.Project`, `install.Global`, CLI orchestration, GC, or build execution.

## Continuation verification

- `go test ./internal/transaction -count=1`: pass.
- `go test -race ./internal/transaction -count=1`: pass.
- `go vet ./internal/transaction`: pass.
- `go test -cover ./internal/transaction -count=1`: pass, 77.7% statement coverage.
- `make check`: pass across the imported product graph.
- `go test -race ./... -count=1`: pass across the imported product graph.
- `make build`: pass on native Darwin/arm64.
- `GOOS=linux GOARCH=amd64 go test -exec=true ./...`: compile graph pass.
- `GOOS=windows GOARCH=amd64 go test -exec=true ./...`: compile graph pass.
- `golangci-lint v2.4.0 run ./internal/transaction/...`: 0 issues.
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file check: clean.

The exact-base task worktree keeps the tracked testing-tool submodule unmaterialized. Commands that traverse the whole repository use the task-local `go.test.mod`, whose semantic difference is an absolute replacement to the canonical submodule checkout. Golangci-lint v2.4.0 rejects `GOFLAGS=-modfile=...` in its internal package loader; the successful scoped lint run omitted that flag because `internal/transaction` does not import the unmaterialized testing-tool module. The lint binary was installed only under the task's ignored `.temp` directory.

Native Darwin subprocess crash-recovery evidence is present. Native Windows runtime execution is unavailable on this Darwin host; Windows cross-compilation is not runtime evidence, and the inherited TASK-260720-1zl1cj native-Windows qualification gate remains preserved.

`task-board validate` still reports the same 13 inherited board-wide issues: 12 legacy EPIC-260712 dependency links and one orphan TASK-260713 resource. None belongs to TASK-260720-31nl14.

Transaction source manifest SHA-256 after rework: `7343c1e72a45e0c0c757c9ed945f2c1944c1ad78b39ab990ce25fde58e52b65f`.
