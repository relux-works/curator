# TASK-260720-31nl14 review cycle 6 verdict

## Verdict

Changes requested. Route to to-dev.

The review-cycle-5 staged-parent fsync fix is present and correct: file and directory staging now syncs the containing directory before PhasePrepared, and sync failure does not publish a prepared journal. One P1 crash-recovery gap remains in PhasePreparing.

## Finding

1. P1 — restart recovery cannot roll back a crash-interrupted partial staged sidecar. Prepare durably writes the PhasePreparing journal before copying each staged target at internal/transaction/engine.go:64-79. The copy then creates the deterministic staged namespace and writes it incrementally: files use OpenFile plus io.Copy at internal/transaction/files.go:81-92, while directories are created and populated recursively at files.go:27-38 and 41-72. A process death or power loss in those intervals can therefore leave a valid durable preparing journal plus a partial regular file or directory whose digest is not DesiredDigest.

Recovery converts PhasePreparing to rollback, pending targets with no backup are untouched, and rollback cleanup passes the staged path with DesiredDigest at internal/transaction/engine.go:505-524, 527-534, and 656-691. removeRecordedSidecar calls validateRemovalStart, which requires the current sidecar digest to equal DesiredDigest at internal/transaction/journal.go:487-517. The expected partial staging state is consequently reported as implementation-corruption and both journal and staging remain, rather than restart deterministically completing rollback.

The current preparing recovery regression does not exercise this boundary: internal/transaction/validation_test.go:280-292 calls copyTarget to successful completion before Recover. The subprocess crash suite covers swap boundaries, not mid-copy preparation.

This violates the acceptance criteria that interrupted work recovers deterministically under the home lock and that fault injection covers target boundaries. It is ordinary implementation rework, not a human-only or external blocker.

## Required rework

- Record enough durable staging ownership/progress before creating or extending a staged sidecar to distinguish transaction-owned partial staging from unknown concurrent state.
- On restart from PhasePreparing, durably remove transaction-owned partial file and directory staging without touching live targets, while still returning implementation-corruption and preserving bytes when the staged namespace has been concurrently replaced or augmented beyond recorded ownership.
- Add subprocess or equivalent crash-boundary regressions for regular-file and nested-directory staging interruption, including partial content/tree recovery and concurrent replacement/addition preservation.
- Preserve the cycle-5 staged-parent durability regression and all prior rollback, cleanup-manifest, namespace, and Windows portability behavior.

## Independent validation

Passed on native Darwin/arm64:

- go test ./internal/transaction -count=1
- go test -race ./internal/transaction -count=1
- go vet ./internal/transaction
- new preparation durability cases repeated 20 times
- core fault, rollback, recovery, cleanup, subprocess, and preparation cases repeated 5 times
- focused statement coverage: 77.3 percent
- make check across the full imported graph with the established task-local module overlay
- go test -race ./... -count=1 across the full imported graph
- native make build and curator --version runtime smoke
- Linux/amd64 and Windows/amd64 complete compile graphs with go test -exec=true ./...
- golangci-lint v2.4.0 scoped to internal/transaction: 0 issues
- gofmt, git diff --check, and staged-file checks clean

No product code was modified or staged during review. Native Windows runtime remains unavailable on this Darwin host; cross-compilation is not runtime evidence, and the inherited TASK-260720-1zl1cj qualification gate remains unchanged.