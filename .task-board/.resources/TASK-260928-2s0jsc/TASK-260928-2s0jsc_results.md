# TASK-260928-2s0jsc — developer results

## Implementation

Added `curator global adopt <command> [--dry-run]`. Adoption resolves the same PATH-visible user-bin directory as global install, requires an existing canonical global target and a regular user-bin file with byte-identical Unix or Windows shim content, and reports the path plus refusal reason for missing, unknown, symlink, special-file, unreadable, or differing entries.

A matching entry is copied under the manager home at `backups/global-bins/` with bytes, permissions, and modification time preserved. The command then records the entry in `.curator-managed.json` using an atomic replacement (POSIX rename; Windows MoveFileEx replace). Real CLI adoption performs a read-only validation before acquiring the manager-home lock, then repeats the checks under the lock. Dry-run and rejected requests do not create lock, backup, or marker state. Repeating a successful adoption is an idempotent no-op.

The install integration test exercises `install.Global` after adoption and confirms it publishes the shim without an unmanaged-conflict warning. Manager-state reads use `internal/stateread`. CLI and troubleshooting documentation cover preview, adoption, and unmanaged-conflict recovery.

## Validation evidence

All commands below ran as standalone processes. Exit codes are the observed process exit codes.

- `go test ./internal/globalbins -count=1` — exit 0. Covers Unix and Windows shim profiles, backup metadata, marker merge, idempotency, dry-run, and refusal cases.
- `go test ./internal/install -run '^TestGlobalInstallRecognizesAdoptedForwardingShim$' -count=1` — exit 0. Exercises the real global install path after adoption.
- `go test ./cmd/curator -run '^TestGlobalAdoptCLI' -count=1` — exit 0. Exercises CLI success, dry-run, mismatch, unknown command, and missing entry.
- `go test ./cmd/curator -run '^TestGlobalAdoptCLIRefusesByteMismatch$' -count=1` with the byte comparison temporarily disabled — exit 1, expected mutant failure: the CLI admitted the different bytes and the refusal assertion failed.
- The same mismatch test after restoring the byte comparison — exit 0.
- `go test ./internal/envprofile -run '^TestManagerOwnedAbsenceReadsAreGuarded$' -count=1` — exit 0.
- `go vet ./internal/globalbins ./internal/install ./cmd/curator` — exit 0.
- `golangci-lint run ./internal/globalbins ./internal/install ./cmd/curator` — exit 0, 0 issues.
- `go build -o /tmp/curator-task-260928-2s0jsc ./cmd/curator` — exit 0.
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/curator-globalbins-task-260928-2s0jsc.test.exe ./internal/globalbins` — exit 0.
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/curator-install-task-260928-2s0jsc.test.exe ./internal/install` — exit 0.
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/curator-cli-task-260928-2s0jsc.test.exe ./cmd/curator` — exit 0.
- `git diff --check` — exit 0.

The local runtime test host was Darwin/amd64. Windows test binaries compile, but Windows runtime execution has not occurred in this producer turn. The hosted gate is triggered by handoff and runs in the runner's finalizing phase; its result is pending and is not claimed here.

The Story worktree preflight reported fresh protected authority for `origin/main`; the selected base and freshly advertised remote OID both matched `97e856425b1aeafe533e86e33e7f9dfd873d9b50`.

## CHANGELOG entry (for release prep)

Add `curator global adopt <command> [--dry-run]` to back up and adopt byte-identical global forwarding shims under Curator management.

No CHANGELOG.md or LOGBOOK.md file was edited. This task-scoped result records the implementation decisions, validation, and remaining platform-evidence limit.
