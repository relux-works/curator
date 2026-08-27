# TASK-260720-256kj1 implementation and rework evidence

## Source boundary

- Detached worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-256kj1/worktree`
- Exact base: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- No files were copied from the dirty shared checkout. Nothing was staged or committed.
- The worktree-local untracked `task-board.config.json` is an orchestration precondition and is not part of the implementation.

## Candidate conformance input

- Candidate root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3ag6pi/worktree/conformance/v1`
- Candidate suite SHA from the handoff: `sha256:70cd274150d629f16ca04f2073a8eae5a0f3fff4584f905d807775040af21eae`
- `vectors/build-drivers.json`: `sha256:fd613bbb4d506237fbabd222247478bc1a809266965f4f4b344ba8548661fe33`
- `expected/build-driver/build-source.preimage.bin`: `sha256:27cdcac0734aa3e069e95a10341e89b118a07c60002516e7b401e95477f01332`
- `expected/build-driver/build-source-sha256.txt`: `sha256:f7155f073664e96cbe66cfe08c33cb87d3ea23ecc80bd889fbbd13567236fbd5`
- This remains candidate evidence only; it is not release or pin evidence.

## Changed implementation surface

- `go.mod`
- `internal/buildsource/buildsource.go`
- `internal/buildsource/buildsource_test.go`
- `internal/buildsource/buildsource_special_unix_test.go`
- `internal/buildsource/conformance_test.go`
- `internal/snapshot/snapshot.go`
- `internal/snapshot/snapshot_test.go`
- `internal/snapshot/lock.go`
- `internal/snapshot/lock_unix.go`
- `internal/snapshot/lock_windows.go`

`internal/hashing` has no diff; legacy `ContentSHA256` semantics and tests remain unchanged.

## Implemented behavior

- Added exact `curator-build-source-v1` validation and uint64-big-endian length-framed identity over every regular file, including root `.csk-install.json`.
- Rejects unsafe protocol paths, platform collisions, duplicate encodings, links, special files, and root/file/tree mutation while ignoring modes and timestamps.
- Retains a validation token through callback use and rechecks the same opened snapshot instance after the last owned child.
- Validates a fresh archive of the immutable repository commit before any snapshot-cache reuse decision.
- Serializes each commit-keyed cache entry with an OS file lock: `flock` on Unix and `LockFileEx` on Windows.
- Preserves the historical `snapshot` path on cold publication. Repair never removes or overwrites a live non-empty directory. It publishes a sibling immutable generation, validates it, and atomically switches a bounded regular-file pointer. The immediately retired generation remains addressable for existing readers; older retired generations are pruned best-effort.
- Repairs changed, missing, extra, linked, marker-tampered, wrong-root-type, and corrupt selection-metadata cases without trusting directory presence.

The generation-pointer design replaces the reviewed target-to-backup sequence. Microsoft documents that `MoveFileEx(..., MOVEFILE_REPLACE_EXISTING)` reports an error when the destination is an existing directory, so a fixed-path directory replacement cannot provide the requested Windows guarantee: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexa

## Review rework evidence

- Added a black-box regression with 24 concurrent mixed `Get` and `GetValidated` repairers plus a live-path observer.
- The pre-fix test reproduced missing-path observations and caller failures.
- The corrected implementation passes 100 repeated stress rounds with zero caller errors, no missing live path, and no staging/backup remnants.
- Added coverage for tampered pointer types, changed active-generation root types, bounded retired generations, and lock serialization.

## Verification

- `CURATOR_CONFORMANCE_ROOT=.../conformance/v1 go test ./internal/buildsource ./internal/snapshot ./internal/hashing -count=1` — pass.
- `go test ./internal/snapshot -run TestConcurrentGetRepairsPresentTamperedSnapshotAtomically -count=100` — pass.
- `go test -race ./internal/buildsource ./internal/snapshot ./internal/hashing -count=1` — pass.
- `go test -cover ./internal/buildsource ./internal/snapshot -count=1` — pass; buildsource 82.0%, snapshot 72.4%.
- `GOOS=windows GOARCH=amd64 go test -c ... ./internal/snapshot` — pass.
- `GOOS=windows GOARCH=amd64 go test -c ... ./internal/buildsource` — pass.
- `GOOS=linux GOARCH=amd64 go test -c ... ./internal/snapshot` — pass.
- `make check` — pass (`go vet ./...`, `go test ./...`, repository gofmt check).
- `go build -o .temp/TASK-260720-256kj1/curator-build ./cmd/curator` — pass.
- `git diff --check` — pass.
- `git diff --exit-code -- internal/hashing` — pass (no diff).
- `golangci-lint` was not run because the binary is unavailable; the repository-required vet/test/gofmt gates pass.

## Known unrelated candidate-suite anomaly

As recorded in the first implementation pass, enabling the candidate conformance root for the entire repository exposes the downstream origin/main `internal/interop/TestManagerLifecycleVectors` reader gap. The task-focused candidate packages and the normal full repository suite pass. This task does not own that downstream reader.

The implementation is ready for review.
