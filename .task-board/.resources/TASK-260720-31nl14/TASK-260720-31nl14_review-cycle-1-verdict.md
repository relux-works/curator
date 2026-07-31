# TASK-260720-31nl14 review cycle 1 verdict

## Verdict

Changes requested. Route to to-dev.

## Findings

1. P1 - Directory desired digests do not bind the root directory metadata, so rollback can overwrite unknown current state. In internal/transaction/digest.go:20-45, DigestPath obtains the root Lstat result but digestDirectory emits only the byte d for the root; modes are encoded only for descendant directories at lines 74-76. A chmod of a committed live directory root therefore leaves DesiredDigest unchanged. rollbackTarget at engine.go:535-618 accepts that changed directory as the desired target, moves it aside, and restores the backup instead of returning ErrImplementationCorruption. cleanupCommitted at engine.go:629-650 can likewise remove the only backup after accepting the wrong root permissions. This conflicts with the desired-digest mismatch and unknown-concurrent-state acceptance criteria. Bind the root directory metadata consistently with child directories, or add an equivalent explicit target-state check, and add rollback plus cleanup tests proving a root metadata change is rejected while the current target is preserved.

2. P1 - Plan and journal validation do not require target namespaces to be independent. buildJournal at engine.go:128-151 rejects only identical cleaned LivePath strings. It accepts ancestor or descendant live paths, path aliases, and a live path that collides with another target sidecar. Because sidecars are placed beside each live path, committing an outer directory can move an inner target and its staging or backup paths before the inner target boundary; such a plan cannot satisfy independent ordered swaps or untouched-target semantics and may be impossible to commit. Reject overlapping or aliased live targets and all cross-target live or sidecar collisions before the first journal write. Add file and directory plan-validation tests, including case-insensitive aliases on native Windows where applicable.

## Validation evidence

- go test ./internal/transaction -count=1: pass.
- go test -race ./internal/transaction -count=1: pass.
- go vet ./internal/transaction: pass.
- make check with the task-local modfile: pass.
- go test -race ./... -count=1 with the task-local modfile: pass.
- Linux amd64 and Windows amd64 complete compile graphs with -exec=true: pass.
- gofmt -l internal/transaction, git diff --check, and staged-file check: clean.
- Native Windows runtime was not available; cross-compilation is not runtime evidence. The inherited TASK-260720-1zl1cj Windows runtime gate remains preserved.

No product code was modified during review.