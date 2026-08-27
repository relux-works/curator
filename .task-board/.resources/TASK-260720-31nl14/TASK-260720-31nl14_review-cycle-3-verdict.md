# TASK-260720-31nl14 review cycle 3 verdict

## Verdict

Changes requested. Route to `to-dev`.

The cycle-2 cleanup-tomb, journal-namespace, and native-Darwin case-alias fixes are present and their regressions pass. Two remaining defects violate independent target boundaries and deterministic crash recovery.

## Findings

1. **P1 — Darwin normalization aliases are accepted as independent target namespaces.** `canonicalNamespacePath` preserves missing component spelling (`namespace.go:95-128`), and `namespaceComponentEqual` applies only `strings.EqualFold` (`namespace.go:173-177`). On the native APFS test filesystem, NFC `é` and NFD `e◌́` names address the same file, but they are not `EqualFold`-equal. The task-scoped probe created two absent live paths using those spellings; it reported `filesystem_alias=true`, while `Engine.Prepare` reported `prepare_accepted=true`. Commit can therefore discover the collision only after mutating an earlier target and entering rollback, contrary to pre-journal namespace independence. Add a native-Darwin normalization-alias regression, conditional on the containing filesystem's observed behavior, and make namespace identity reflect the filesystem's actual component equivalence before journal creation.

2. **P1 — A crash during recursive directory-tomb deletion cannot be recovered.** `removeDurably` renames a recorded directory to `.delete`, then calls `os.RemoveAll` (`journal.go:376-430`). If the process dies after `RemoveAll` has deleted only part of that directory, restart takes the tomb branch and requires the partial tree to retain the original full digest (`journal.go:397-409`). The task-scoped probe paused a committed directory transaction in `PhaseCleanup`, renamed the backup to its canonical tomb, removed one tomb entry to model interruption inside `RemoveAll`, and restarted recovery. Recovery returned `ErrImplementationCorruption` and retained the partial tomb/journal instead of completing cleanup. This state is an ordinary crash window, not unknown concurrent state. Make per-removal cleanup progress durably distinguish a transaction-owned in-progress tomb from a preexisting/unknown tomb, or retain sufficient durable provenance to validate a partial remainder. Add direct and restart fault injection after tomb rename and during directory removal; recovery must finish cleanup while existing unknown/simultaneous-tomb regressions continue to preserve foreign bytes and return implementation-corruption.

## Validation evidence

Passed independently on native Darwin/arm64:

- `go test ./internal/transaction -count=1 -v` (including subprocess crash recovery and exhaustive target-boundary injection)
- `go test -race ./internal/transaction -count=1`
- `go vet ./internal/transaction`
- `go test -cover ./internal/transaction -count=1` — 77.8% statement coverage
- `golangci-lint v2.4.0 run ./internal/transaction/...` — 0 issues
- `make check`
- `go test -race ./... -count=1`
- complete Linux/amd64 and Windows/amd64 compile graphs via `go test -exec=true ./...`
- `gofmt`, `git diff --check`, and staged-file checks

No product code was modified or staged during review. Native Windows runtime remains unavailable; Windows compilation is not runtime evidence, and the inherited `TASK-260720-1zl1cj` final qualification gate must remain preserved after rework.

## Task-scoped probes

- `.temp/TASK-260720-31nl14/worktree/.temp/review-cycle-3/unicode-alias-probe.go`
- `.temp/TASK-260720-31nl14/worktree/.temp/review-cycle-3/cleanup-crash-probe.go`
- `.temp/TASK-260720-31nl14/review-cycle-3-tool-readiness.txt`
