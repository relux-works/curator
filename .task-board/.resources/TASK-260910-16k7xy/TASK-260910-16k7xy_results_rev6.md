# TASK-260910-16k7xy rev6 results — Windows identity-completion fix on top of rev5

Base: story checkpoint a4181f9 (14hsti) on task-board/story/STORY-260910-24nyb1.
Worktree carries only the 6 leaf paths, uncommitted (no commits by this run):
M internal/envprofile/boundaries_test.go, M internal/envprofile/envprofile.go,
M internal/snapshot/boundaries.go, ?? internal/envprofile/draft_capture_test.go,
?? internal/snapshot/capture.go, ?? internal/snapshot/capture_test.go.
Rework-2 (P1-a, P1-b, P2) content from rev5 is intact; this revision adds one
minimal Windows-correctness fix (no other changes).

## Root cause of the rev5 Windows gate failure

Run 35165167680 (rev5): Test windows-latest failed `go test` (exit 1,
platform-case gate ok). Windows evidence artifact
test-evidence-windows-latest, go-test-served.json: exactly one failing test —
`FAIL github.com/relux-works/curator/internal/snapshot ::
TestReviewIdentityReplacementDuringCapture`
(`capture_test.go:374: Capture err = <nil>, want leading
source_snapshot_changed`).

Cause: rev5 `statIdentities` stored raw `os.Lstat` records. On Windows the
volume serial + file index that `os.SameFile` compares are resolved lazily
from the stored path on first `SameFile` call (Go src/os/types_windows.go:
`loadFileId` caches vol/idx and clears path on first use). The baseline
record was therefore resolved AFTER the hook's rename-replace, against the
new file — baseline and current compared as the same object and the
same-byte replacement passed. Fix (internal/snapshot/capture.go:490):
complete each identity eagerly at inventory time with
`os.SameFile(info, info)` (the established in-repo pattern from
internal/transaction/namespace_identity_windows.go `completeNamespaceIdentity`);
an unreadable identity fails closed (`cannot read file identity`). On unix
this is a no-op (dev/ino are eager). No new skip class; the rev4
platform-control skips are untouched.

## Evidence (shell: sh/zsh, exit codes are the command's own, no pipes)

- gofmt -l internal/snapshot/ internal/envprofile/ : exit 0, no output.
- go vet ./internal/snapshot/ ./internal/envprofile/ : exit 0.
- GOOS=windows go vet ./internal/snapshot/ : exit 0.
- GOOS=windows go test -c -o /dev/null ./internal/snapshot/ : exit 0 (compile).
- go test -p 1 -count=1 ./internal/snapshot/ : exit 0, ok 1.466s (full package).
- go test -p 1 -count=1 -v ./internal/snapshot/ -run
  TestReviewIdentityReplacementDuringCapture|TestCaptureDetectsLiveContentChange|
  TestPublishRefusesCopyMutatedAfterAudit|TestCaptureMatchesNormativeVectors :
  exit 0; all PASS (incl. all 3 normative vector subtests base/runtime-edit/
  build-edit on darwin).
- go test -p 1 -count=1 ./internal/envprofile/ -run TestDraft : exit 0,
  ok 11.083s.
- golangci-lint run ./internal/snapshot/... : exit 0, 0 issues.
- golangci-lint run ./internal/envprofile/... : exit 0, 0 issues.

## Mutant kill (identity gate; applied, tested, restored byte-identical)

- Narrowing `if !ok || !os.SameFile(prior, info)` ->
  `if !ok || (false && os.SameFile(prior, info))`:
  go test -p 1 -count=1 ./internal/snapshot/ -run
  TestReviewIdentityReplacementDuringCapture : exit 1,
  `capture_test.go:374: Capture err = <nil>, want leading
  source_snapshot_changed`, FAIL. Killed.
- File restored afterwards; `diff -q` confirms identical bytes and the
  follow-up full-package run above is green (exit 0).

## Bounds and notes

- Windows run-proof of the fix itself comes from the hosted gate on handoff
  (no Windows runner on this host); cross-compile + vet pass locally.
  Reasoning is pinned to Go source (lazy `loadFileId`) and the repo's own
  Windows identity precedent, not to a proxy signal.
- Production wiring unchanged from rev5: Install/stateForPath ->
  admitPathSource -> PrepareLocalAcquisition -> Capture -> PublishLocal ->
  OpenLocal -> contextstore.EnsureState; draft-only behind DraftSourcesV1;
  frozen v1 untouched; no synthesized commits; no live-directory consumption
  after capture.

## Checklist

- [x] Rework-2 P1-a/P1-b/P2 retained; Windows identity-completion fix added
- [x] Rev5 Windows failure root-caused to lazy SameFile with CI evidence
- [x] gofmt/vet (incl. windows)/lint clean; narrow package tests green
- [x] Identity narrowing mutant killed (exit 1), file restored, green re-run
- [x] Worktree holds only the 6 leaf paths, uncommitted
