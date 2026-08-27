# TASK-260720-31nl14 review-cycle 3 rework validation

## Resolved findings

- Darwin namespace identity now normalizes canonical Unicode spellings on APFS and HFS before component comparison. The native APFS probe reports filesystem_alias=true and prepare_accepted=false, so NFC/NFD aliases are rejected before journal creation.
- Canonical journals now record removal_path durably before a sidecar-to-tomb rename. Cleanup uses deterministic recursively durable deletion and retains ownership progress until the tomb is durably absent.
- Public fault boundaries after_cleanup_rename and during_cleanup_removal prove direct and restart recovery for full and partial directory tombs in both commit cleanup and rollback cleanup. Existing unowned and simultaneous tomb tests still return implementation-corruption without deleting foreign bytes.

## Verification

- Focused transaction tests: pass.
- Transaction race: pass.
- Transaction vet: pass.
- Transaction coverage: 78.2 percent.
- Repository make check: pass.
- Repository full race: pass.
- Native Darwin arm64 build and version smoke: pass.
- Linux amd64 complete compile graph: pass.
- Windows amd64 complete compile graph: pass.
- golangci-lint v2.4.0 scoped to internal/transaction: 0 issues.
- gofmt, git diff check, and staged-file check: clean.

The work remains isolated to internal/transaction in the task-scoped exact-base worktree. No files were staged or committed. Native Windows runtime execution is unavailable on this Darwin host; cross-compilation is not runtime evidence, and the inherited TASK-260720-1zl1cj Windows qualification gate remains preserved. task-board validate still reports 13 inherited board-wide issues: 12 legacy EPIC-260712 broken links and one orphan TASK-260713-7a9c1e resource; none belongs to this task.