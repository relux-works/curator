# TASK-260910-16k7xy rev6 results (rework-3: Windows identity fix)

Workspace: STORY-260910-24nyb1 worktree, branch task-board/story/STORY-260910-24nyb1, checkpoint c79d0fc (14hsti) + rev5 patch + this fix, uncommitted.

## Change vs rev5 (only change)
Windows TestReviewIdentityReplacementDuringCapture failed on the gate with `Capture err = <nil>, want leading source_snapshot_changed`: os.SameFile on Windows resolves the file index lazily from the path (FileInfo carries no public file index), so an atomic same-byte replacement was invisible. Fix per rework-3:
- NEW internal/snapshot/identity_windows.go (windows build): captureFileToken delegates to the checkpointed staging.FileIdentity primitive (single backup-semantics open, volume serial + file index).
- NEW internal/snapshot/identity_unix.go (unix build): captureFileToken returns dev:ino from a single Lstat, identical semantics to the previous os.SameFile comparison.
- internal/snapshot/capture.go: statIdentities/revalidateIdentities now pin and compare eager string tokens instead of lazily-resolved os.FileInfo; doc comments updated. Unix detection outcome unchanged.

## Evidence (macOS arm64, go1.26.0; no local Windows, gate is the Windows proof)
- gofmt -l internal/snapshot/ internal/envprofile/ -> clean, exit 0
- go vet ./internal/snapshot/ ./internal/envprofile/ -> clean, exit 0
- GOOS=windows go build ./internal/snapshot/ ./internal/envprofile/ ./internal/staging/ -> exit 0
- GOOS=windows go vet ./internal/snapshot/ -> exit 0
- go test -p 1 ./internal/snapshot/ -run TestCapture|TestVectors|TestReview|TestPublish|TestOpen -count=1 -v -> all 17 PASS incl. 3 normative vectors and TestReviewIdentityReplacementDuringCapture, exit 0
- go test -p 1 ./internal/snapshot/ -count=1 (full package) -> ok, exit 0
- go test -p 1 ./internal/envprofile/ -run TestDraft -count=1 -v -> all 7 PASS incl. TestDraftInstallFreezesLiveMutations, exit 0
- golangci-lint run ./internal/snapshot/... -> 0 issues, exit 0
- Mutant probe: captureFileToken stubbed to a constant -> TestReviewIdentityReplacementDuringCapture FAILS with the exact rev5 Windows gate signature (exit 1); reverted -> PASS (exit 0).

## Checklist
- [x] rev5 patch applied; workspace carries only rev5 paths + 2 new identity files
- [x] Windows identity uses staging single-handle primitive; unix behavior unchanged
- [x] Narrow tests green, lint clean, cross-compile clean
- [x] Mutant killed with gate-matching signature
- [x] Work left uncommitted for handoff snapshot
