# TASK-260720-31nl14 review cycle 5 verdict

## Verdict

Changes requested. Route to to-dev.

The cycle-4 per-entry cleanup manifest fix is present and its direct/restart, commit/rollback, added/replaced-state regressions pass. One P1 crash-durability defect remains before the first target swap.

## Finding

1. P1 — the staged sidecar namespace entry is not made durable before the journal advances to prepared. For a regular target, copyTarget returns directly from copyRegular (internal/transaction/files.go:21-23), which syncs the file contents but never syncs filepath.Dir(destination) (files.go:75-105). For a directory target, copyTarget returns syncTree(destination) (files.go:27-38); syncTree syncs the destination tree and root directory itself, but not the parent directory containing the newly created destination entry (files.go:108-132). Prepare then verifies the digest and durably changes the journal to PhasePrepared without any parent-directory sync (internal/transaction/engine.go:71-90), although PhasePrepared is defined as fully durable staging (internal/transaction/types.go:38-43).

On filesystems where file or child-directory fsync does not persist the containing directory entry, a power loss can therefore leave a durable prepared journal while losing the staged sidecar name. Restart then cannot deterministically resume the recorded commit and instead falls into an avoidable commit failure/rollback. This violates the durable-before-next-swap and restart-recovery acceptance criteria.

## Required rework

- Make creation of both regular-file and directory staged sidecars durable by syncing the containing parent directory after the complete staged tree and digest are durable, and before saving PhasePrepared.
- Add a fault-injectable or observable regression for file and directory preparation proving the staged parent-directory durability primitive succeeds before the prepared journal transition. A durability-primitive failure must not expose a prepared journal and must retain deterministic rollback/recovery state.
- Preserve the cycle-4 cleanup manifest regressions and all existing platform-specific durability behavior.

This is ordinary implementation rework, not a stop-the-line or human-only decision.

## Independent validation

Passed on native Darwin/arm64:

- go test ./internal/transaction -count=1
- go test -race ./internal/transaction -count=1
- go vet ./internal/transaction
- focused cleanup/subprocess regressions repeated five times
- 77.3% focused statement coverage
- make check
- go test -race ./... -count=1
- native make build and curator dev runtime smoke
- Linux/amd64 and Windows/amd64 complete compile graphs via go test -exec=true ./...
- golangci-lint v2.4.0: 0 issues
- gofmt, git diff --check, and staged-file checks

No product code was modified or staged during review. The generated native build binary was moved to Trash after the runtime smoke. Native Windows execution is unavailable on this Darwin host; Windows compilation is not runtime evidence, and the inherited TASK-260720-1zl1cj qualification gate remains preserved.