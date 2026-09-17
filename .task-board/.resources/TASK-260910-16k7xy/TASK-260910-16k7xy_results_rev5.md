# TASK-260910-16k7xy rev5 results — rework 2 (verdict rev4 findings)

Base: story checkpoint a4181f9 (14hsti) on task-board/story/STORY-260910-24nyb1.
Worktree carries only the 6 leaf paths (3 tracked-modified + 3 new):
M internal/envprofile/boundaries_test.go, M internal/envprofile/envprofile.go,
M internal/snapshot/boundaries.go, ?? internal/envprofile/draft_capture_test.go,
?? internal/snapshot/capture.go, ?? internal/snapshot/capture_test.go.
No commits made by this run (handoff snapshots the working tree).

## What changed (nothing else)

P1-a — internal/snapshot/capture.go: Capture now records the physical identity
of every admitted file at inventory time (statIdentities, os.Lstat) and
revalidates it after the copy loop (revalidateIdentities via os.SameFile:
device+inode on unix, volume+file-index on Windows). A replaced file with
identical bytes and mode refuses with source_snapshot_changed. Deterministic
seam: captureAfterCopyHook (test-only, nil in production).

P1-b — internal/snapshot/capture.go: PublishLocal now hashes the actual
publication copy (tmp tree) immediately before the rename and refuses with
source_snapshot_changed when it diverges from the audited digest; the deferred
tmp removal leaves no store entry behind. Deterministic seam:
publishAfterAuditHook (test-only, nil in production).

P2 — internal/envprofile/envprofile.go: admitPathSource resolves through
OpenLocal, then runs admitPathSourceAfterFreezeHook (test-only, nil in
production) before returning the frozen tree to stateForPath/EnsureState.
internal/envprofile/draft_capture_test.go: TestDraftInstallFreezesLiveMutations
now mutates the live directory AFTER capture and BEFORE contextstore
consumption inside ONE Install and asserts the installed bytes are the frozen
ones (plus a sanity read that the live mutation really landed).
internal/snapshot/capture_test.go: TestCaptureDetectsLiveContentChange is now a
race within one Capture (hook rewrites the live file after the copy loop);
added TestReviewIdentityReplacementDuringCapture (reviewer name verbatim) and
TestPublishRefusesCopyMutatedAfterAudit.

## Evidence (shell: zsh with set -o pipefail where piped; exit codes real)

- gofmt -l internal/snapshot/ internal/envprofile/ : exit 0, no output.
- go vet ./internal/snapshot/ ./internal/envprofile/ : exit 0.
- golangci-lint run ./internal/snapshot/... : exit 0, 0 issues.
- golangci-lint run ./internal/envprofile/... : exit 0, 0 issues.
- go test -p 1 -count=1 ./internal/snapshot/ : exit 0, ok 4.396s (full package).
- go test -p 1 -count=1 -v ./internal/snapshot/ -run
  TestCaptureDetectsLiveContentChange|TestReviewIdentityReplacementDuringCapture|
  TestPublishRefusesCopyMutatedAfterAudit|TestCaptureMatchesNormativeVectors :
  exit 0; all PASS incl. all 3 normative vector subtests (base, runtime-edit,
  build-edit) on darwin.
- go test -p 1 -count=1 ./internal/envprofile/ -run TestDraft : exit 0,
  ok 13.022s.
- go test -p 1 -count=1 -v ./internal/envprofile/ -run
  TestDraftInstallFreezesLiveMutations|TestDraftInstallAdmitsDirtyUntrackedInsideGit|
  TestDraftInstallRefusesSnapshotInsideStore : exit 0; all 3 PASS.

## Mutant kills (each applied, tested, then restored byte-identical via diff -q)

- Identity narrowing (|| SameFile -> &&, refuses nothing): 
  TestReviewIdentityReplacementDuringCapture FAILS, go test exit 1
  (Capture err = nil, want leading source_snapshot_changed). Killed.
- Pre-rename hash skip (if false && ...): 
  TestPublishRefusesCopyMutatedAfterAudit FAILS, go test exit 1
  (PublishLocal err = nil, want leading source_snapshot_changed). Killed.
- Reviewer capture narrowing (digest compare -> file-count compare):
  TestCaptureDetectsLiveContentChange FAILS, go test exit 1
  (Capture err = nil, want leading source_snapshot_changed). Killed.
- Reviewer install narrowing (source = stored -> only when stored == ...):
  TestDraftInstallFreezesLiveMutations FAILS, go test exit 1
  (installed a.md = mutated, want frozen acme bytes). Killed.

Narrowing mutant score on the reviewer attacks: killed 2/2 (plus 2/2 new
gate narrowings). Every mutant file restored afterwards; diff -q confirms
rev5 bytes back in place before the green runs above.

## Bounds and notes

- Windows: the rev4 platform-control skips are untouched
  (gated on GOOS==windows, reason starts with the declared sentence
  Windows does not expose portable executable permission bits). The new
  identity path uses os.SameFile, which compares volume+file-index on
  Windows, so the committed identity race test is expressible there; no new
  skip class added. Windows/CI evidence comes from the hosted gate on handoff.
- Production wiring unchanged: Install/stateForPath -> admitPathSource ->
  PrepareLocalAcquisition -> Capture -> PublishLocal -> OpenLocal ->
  contextstore.EnsureState; draft-only behind DraftSourcesV1; frozen v1
  untouched; no synthesized commits; no live-directory consumption after
  capture. Mutant probes above used only local file copies under /tmp for
  backup/restore; no board bytes edited by hand.

## Checklist

- [x] P1-a identity revalidation + committed reviewer test
- [x] P1-b pre-rename publication hash + committed refusal test, no leftover entry
- [x] P2 in-Install freeze proof + single-capture race test
- [x] Reviewer mutants killed (2/2), new narrowings killed (2/2)
- [x] gofmt/vet/lint clean; narrow package tests green with exit codes
- [x] Worktree holds only the 6 leaf paths, uncommitted
