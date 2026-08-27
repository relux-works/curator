# TASK-260720-1ljev5 review verdict — cycle 3

## Verdict: ACCEPTED

Route to `done`. No task-owned blocking or rework finding remains.

## Review result

The final task tree closes both P1 findings from cycle 2 and satisfies the GC
acceptance criteria:

- `internal/scopes/consumers.go` parses the registry as a duplicate-aware token
  stream. Repeated object members, ambiguous/noncanonical versions, unsupported
  members, malformed consumer lists, and trailing content fail closed before
  `Collect`, `RecordConsumer`, or `StageConsumer` can normalize the file.
  Consecutive-pass regressions preserve the original registry bytes and keep
  the live marker reference discoverable after repair.
- `internal/buildcache/collect.go` and
  `internal/buildcache/protection_unix.go` keep decisive classification on the
  proven cache-entry descriptor. Receipt, exact members, artifact directory,
  artifact bytes/hash/size, and publication metadata are obtained through held
  handles; Unix uses `openat` plus `O_NOFOLLOW`. The sweep does not reopen the
  manager home, cache-root pathname, or candidate pathname after proof.
- The Windows implementation holds the protected boundary without
  `FILE_SHARE_DELETE`, rejects reparse points and untrusted owner/DACL state,
  revalidates entry identity, and reads the decisive entry listing from the
  held handle.
- Marker v2 keys from project/global/hybrid scopes and transaction journal keys
  feed the same locked mark/sweep pass. Uncertain registry/scope/marker state
  blocks build sweeping and survives across consecutive passes. Consumer
  pruning, runtime GC, journal recovery, install commit/rollback, and standalone
  GC remain serialized on the manager-home lock.
- Retirement remains handle-root-bound through `os.Root`; root exchange,
  symlink/junction/reparse, corrupt receipt, untrusted provenance, grace,
  interrupted-removal, and direct-child containment cases are covered and
  green.

## Independent validation

Reviewer-run on darwin/arm64 with Go 1.25.5:

- PASS: `go test ./internal/scopes ./internal/buildcache ./cmd/curator -count=1`
- PASS: `go test -race ./internal/scopes ./internal/buildcache ./cmd/curator -count=1`
- PASS: focused install maintenance/concurrency race suite
- PASS: `gofmt -l .`, `git diff --check`, `go vet ./...`, `go build ./...`
- PASS: `go test ./... -count=1`, including `internal/install` (279.776s) and
  `internal/install/atomicity` (395.597s)
- PASS: Windows and Linux cross-builds; PASS: Linux cross-vet
- Expected inherited red: full Windows cross-vet reports only
  `internal/runtimestore/targets_windows_test.go:97:14: undefined:
  decodeHelperOutput`; the untouched accepted base reproduces the identical
  output.

The reviewer shell did not contain `golangci-lint`. The attached cycle-3 gate
artifact was therefore inspected directly: its lint log records `0 issues`
with exit 0. Both expected-red negative controls fail for the intended defect:
restoring struct decoding loses the repeated-registry protections, and
restoring pathname classification retires an entry on another object's proof.

Native Windows evidence was inspected case by case. Task-owned scopes, GC,
registry, junction/reparse, root-exchange, classification, containment, and
sweep tests execute and pass. Remaining Windows failures reproduce on the
accepted base with the same DACL-inheritance race or are the previously accepted
file-reparse privilege limitation; none is introduced by this task delta.

No product code was modified during review.
