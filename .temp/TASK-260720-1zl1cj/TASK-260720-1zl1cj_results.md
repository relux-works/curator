# TASK-260720-1zl1cj implementation evidence

## Provenance

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-1zl1cj/worktree`
- Detached baseline: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` (`origin/main`)
- Imported accepted source: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-4bd0it/worktree`
- Accepted and imported tracked binary-diff SHA-256: `ed23bdd5db679114bcdb4ea53134f84d602d0bbf6c80820aef61f653fc66a342`
- All 44 accepted untracked product files are byte-identical after excluding task-local `internal/managerlock`; board/config/temp/planning/research/diagram/binary/cache files were not imported.
- No files were staged or committed.

## Implementation

Added `internal/managerlock` with:

- canonical absolute project identities, existing-symlink resolution, duplicate rejection, and unsigned UTF-8 byte ordering;
- stable SHA-256-derived project, logical build-key, and manager-home lock identities below `<curator-home>/state/locks/v1`;
- an operation state machine enforcing ordered project batches, at most one optional build-key lock, key release before home acquisition, reverse release order, and home-only recovery/GC acquisition;
- a process-local reservation guard that rejects project/key acquisition while a home lock is held or pending and rejects home acquisition while a key is held or pending;
- explicit dry-run acquisition rejection before filesystem I/O;
- context-cancellable nonblocking lock acquisition using `flock` on Unix and `LockFileEx` on Windows;
- stable lock files that are never deleted on release, while OS locks are released by close or abnormal process exit;
- subprocess coverage for same/different project locks, manager-home contention, build-key deduplication across independent projects, cancellation rollback, and abnormal child exit.

## Validation

- Native platform: Darwin arm64, Go 1.25.5.
- `go test -race -cover ./internal/managerlock -count=1 -v`: pass, 82.1% statement coverage.
- `go test ./internal/managerlock -count=5`: pass.
- `go vet ./internal/managerlock`: pass.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.4.0 run ./internal/managerlock/...`: 0 issues.
- `make check`: pass after materializing the repository's pinned `agents` submodule in the detached worktree.
- `go test -race ./... -count=1`: pass.
- Native `go build ./cmd/curator`: pass; generated Mach-O arm64 smoke binary outside the worktree.
- Linux amd64 managerlock test binary: compiled successfully (`ELF 64-bit x86-64`).
- Windows amd64 managerlock test binary: compiled successfully (`PE32+ x86-64`).
- Native Windows runtime was unavailable on this Darwin host (no Wine, Docker, or Podman), so the platform-neutral subprocess suite was run natively on Unix and Windows runtime execution is deferred to Windows CI/review infrastructure.
- `gofmt -l internal/managerlock`: empty; `git diff --check`: pass.

## Lint boundary finding

Full-repository golangci-lint reports 64 inherited issues in the reviewer-accepted imported buildcache/buildsource/godriver/runtimestore/snapshot and related code. None are in `internal/managerlock`; the task-scoped lint gate reports zero issues. Those unrelated accepted files were deliberately not modified under this task's package ownership constraint.
