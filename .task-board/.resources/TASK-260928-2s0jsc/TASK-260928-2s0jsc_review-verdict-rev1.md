# TASK-260928-2s0jsc review verdict — rev1 ACCEPTED (reviewer claude-opus-5-5 low)

Candidate: base 97e85642, tree 40b10737; worktree files hash-identical to the candidate tree (5 untracked blobs checked via git hash-object; tracked diff vs tree empty).

## Verified
1. adopt.go: name must be identifiers.Valid AND canonical target `<home>/global/bin/<name>` must exist as a regular file; entry via stateread.Lstat (symlink / non-regular refused); bytes compared with runtimestore.UnixShimContent / WindowsShimContent(canonical); entry never rewritten; backup under `<home>/backups/global-bins` (perm + mtime preserved, cp -p equivalent); marker via writeLedger temp+rename (Unix) / MoveFileEx REPLACE_EXISTING|WRITE_THROUGH (Windows) — atomic.
2. Refusals return before any write; ledger read via stateread.ReadRegularFile (unreadable/malformed ≠ absent, UnusableError).
3. dry-run returns before writes; already-managed short-circuits (idempotent); internal/install TestGlobalInstallRecognizesAdoptedForwardingShim drives global install afterwards.
4. CLI: read-only preflight, then AcquireHomeOnly (same manager-home lock global install uses, commit.go:180) and full re-check under lock; entry re-confirmed (SameFile, mode, mtime, bytes) before and after backup. Crash between backup and marker leaves an orphan backup only — marker is unchanged/consistent.
5. Docs: cli.md + troubleshooting.md updated. No CHANGELOG/LOGBOOK, no stray files.

## Local runs (bash, pipefail, real exit codes)
- go test ./internal/globalbins ./internal/install -run "Adopt|Ledger|Global" → rc=0
- go test ./cmd/curator -run GlobalAdopt → rc=0
- go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded → rc=0
- GOOS=windows go vet ./internal/globalbins ./cmd/curator → rc=0

## Mutants (adopt.go, restored byte-identical after each)
- M1 byte comparison skipped → KILLED (globalbins TestAdoptRefusalsDoNotWrite + CLI TestGlobalAdoptCLIRefusesByteMismatch), rc=1
- M2 lstat→stat for the entry → KILLED (TestAdoptRefusalsDoNotWrite "symbolic link"), rc=1
- M3 marker written on refusal → KILLED (both packages), rc=1
- M4 dry-run gate removed → KILLED (dry-run test + CLI idempotency test), rc=1

## Bounds / residuals (non-blocking)
- R1: symlink refusal is killed only at the globalbins.Adopt level; the CLI suite survives M2 (no CLI symlink row). Adopt is the sole production call path, so bound is small.
- R2: "Curator global command" is established by existence of the canonical target in `<home>/global/bin`, not by the config's global command set — a stale canonical target from a removed command could be adopted. Reasonable reading of "canonical target"; note for follow-up.
- R3: Windows bytes checked cross-platform on darwin (WindowsShimContent); real Windows behaviour (MoveFileEx) only via the hosted windows lane (gate reported green by orchestrator; not re-downloaded here).
