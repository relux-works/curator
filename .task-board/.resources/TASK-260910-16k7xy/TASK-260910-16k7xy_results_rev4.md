# TASK-260910-16k7xy results — rev4 re-apply (final leaf of STORY-260910-24nyb1)

Re-applied the rev4 candidate patch (`git apply --binary`) onto current trunk
checkpoint `a4181f9` after the CR construction refused
`change_request_base_authority_mismatch`. No other changes.

## Workspace state

- Story branch: `task-board/story/STORY-260910-24nyb1`, HEAD `a4181f9b86158f67bf4a447794334f999a7846d7`
- `git status --short` shows exactly the 6 candidate paths:
  - `M internal/envprofile/boundaries_test.go`
  - `M internal/envprofile/envprofile.go`
  - `M internal/snapshot/boundaries.go`
  - `?? internal/envprofile/draft_capture_test.go`
  - `?? internal/snapshot/capture.go`
  - `?? internal/snapshot/capture_test.go`
- Work left uncommitted for the handoff snapshot.

## rev4 content (the only two gate defects, both fixed)

1. Lint (`revive indent-error-flow`): `internal/snapshot/capture.go` — the
   `else` after the returning `if` in `PublishLocal` is gone; the conflict
   error returns after the `if` block.
2. Windows platform-case gate: `internal/snapshot/capture_test.go:105,113` —
   both skips start with the declared platform-control reason verbatim,
   "Windows does not expose portable executable permission bits", gated on
   `runtime.GOOS == "windows"` only. On unix a missing execute bit is a
   `t.Fatalf` (test failure), not a skip. No skip class added.
   Declared deferral: the normative executable vectors are asserted on the
   unix runners; on Windows the permission bits are synthesized so the
   executable-flag assertion is deferred there (platform-control class).

## Evidence (all run locally in the Story worktree, exit codes quoted)

- `go vet ./internal/snapshot ./internal/envprofile` → exit 0
- `go test -p 1 ./internal/snapshot -run 'Capture|Vector' -count=1` → exit 0
  (`ok ... 5.175s`; rerun `-v` → exit 0, `ok ... 2.816s`):
  - `TestCaptureMatchesNormativeVectors/{base,runtime-edit,build-edit}` PASS (all three run, none skipped on darwin)
  - `TestVectorsFreezeContextAcrossRuntimeAndBuildEdits` PASS
  - `TestCaptureAdmitsDirtyAndUntrackedInsideGit` PASS
  - `TestCaptureDetectsAddedMember`, `TestCaptureDetectsRemovedMember` PASS
  - `TestCaptureDetectsLiveContentChange` PASS
  - `TestCaptureRefusesUnprepared` PASS
  - `TestCaptureDetectsLinkSwap` PASS
- `go test -p 1 ./internal/envprofile -run 'TestDraftInstall' -count=1 -v` → exit 0
  (`ok ... 19.742s`):
  - `TestDraftInstallFreezesLiveMutations` PASS
  - `TestDraftInstallAdmitsDirtyUntrackedInsideGit` PASS
  - `TestDraftInstallRefusesSnapshotInsideStore` PASS
- `golangci-lint run ./internal/snapshot/...` → `0 issues.`, exit 0
- `golangci-lint run ./internal/envprofile/...` → `0 issues.`, exit 0
- `gofmt -l` on all 6 touched files → no output (clean), exit 0
- Full `./internal/envprofile` suite (no `-run` mask) deliberately not run:
  host stalls on new test binaries; per wave note, narrow package tests only.
  (One unmasked run was started and terminated before completion; not claimed.)

## Checklist

- [x] Rev4 candidate applied, exactly 6 paths, no other changes
- [x] Lint clean locally (`golangci-lint`, `gofmt`, `go vet`)
- [x] Snapshot capture/vector tests pass, incl. all three normative vectors
- [x] Draft install production-path tests pass (freeze, dirty/untracked, refusal)
- [x] Windows skips use the declared platform-control reason verbatim, windows-gated only
- [x] Work uncommitted; ready for `task-board handoff TASK-260910-16k7xy --role developer`
