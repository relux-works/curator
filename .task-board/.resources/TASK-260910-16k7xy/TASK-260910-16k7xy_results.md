# TASK-260910-16k7xy results — capture and store local package snapshots (developer handoff evidence)

Worktree: STORY-260910-24nyb1 (branch task-board/story/STORY-260910-24nyb1), HEAD 3dbefcd.
Contract: curator-spec main protocol/skillfile-sources.md + repository-transport.md (rev-1 semantics),
schemas/draft-sources-v1/local-snapshot-v1.schema.json,
conformance/draft-sources-v1/snapshot-cases.json. Work left UNCOMMITTED for the handoff snapshot.

## What this run did

The worktree already carried a prior uncommitted capture implementation (capture.go,
capture_test.go, draft_capture_test.go, wiring in envprofile.go + snapshot/boundaries.go).
This run audited it against the accepted contract, closed three refusal-coverage gaps with new
negative tests, and verified everything below. No unrelated dirty work was present; all dirty
files belong to this task.

Production path: `admitPathSource` (internal/envprofile/envprofile.go) runs
`snapshot.PrepareLocalAcquisition` (boundaries) -> `snapshot.Capture` (copy admitted
dirty/untracked bytes incl. inside Git into private staging, hash, revalidate) ->
`snapshot.PublishLocal` (rehash + atomic renameNoReplace into manager-home `local-snapshots/`
store keyed by digest) -> `snapshot.OpenLocal` (authenticate, never recreate). `stateForPath`
installs the frozen snapshot dir via `contextstore.EnsureState`, never the live directory.
Path overlays route through the same `stateForPath`, so both entries are covered. Draft-only
behind `DraftSourcesV1`; frozen v1 path (live-dir traversal) unchanged.

## Acceptance mapping (each row reaches a production entry point)

- Dirty/untracked inside Git captured, `.git` excluded, no commit synthesized:
  `TestCaptureAdmitsDirtyAndUntrackedInsideGit` (snapshot level, asserts `git rev-list --count HEAD == 1`)
  + `TestDraftInstallAdmitsDirtyUntrackedInsideGit` (Install level, installed bytes are dirty, not HEAD).
- Exact byte inventory + executable flags, all three normative vectors:
  `TestCaptureMatchesNormativeVectors` pins per-file sha256, exec bits and snapshot digests
  base / runtime-edit / build-edit against conformance/draft-sources-v1/snapshot-cases.json
  (verified byte-identical to the spec file this run). Independent oracle: system sha256sum of the
  three base file contents reproduces 06142a8f…, 0cb42bbd…, df1d036c… exactly.
- Runtime/build/dependency inputs frozen: `TestVectorsFreezeContextAcrossRuntimeAndBuildEdits`
  (SKILL.md bytes identical across vectors, three distinct digests) — script-only and
  build-only changes change package identity per spec section 3.
- Capture races fail: `TestCaptureDetectsAddedMember`, `TestCaptureDetectsRemovedMember`
  (membership), `TestPublishDetectsFrozenMutation` (frozen rehash), `TestCaptureDetectsLinkSwap`
  (link smuggled post-admission; NEW this run).
- Missing snapshots fail, never recreated: `TestOpenMissingSnapshot`
  (`source_snapshot_unavailable`, incl. after store removal), `TestOpenMalformedDigest` (NEW).
- Tampered store refused: `TestOpenTamperedSnapshot` (`source_snapshot_changed`).
- Unprepared calls refused: `TestCaptureRefusesUnprepared` (`Capture(nil)`, empty acquisition,
  digest-less `PublishLocal`; NEW; asserts no store state left behind).
- Live dir never used after capture: `TestDraftInstallFreezesLiveMutations` (mutate live tree
  post-Install, installed entry bytes unchanged; exactly one `local-snapshots/` entry).
- Snapshot store is a managed output: `TestDraftInstallRefusesSnapshotInsideStore`
  (`source_output_overlap`, publishes nothing).

## Verification (narrow only, shell bash, `set -o pipefail` where exit recorded)

- `go test -count=1 ./internal/snapshot/` -> ok, exit 0 (16 tests incl. 3 new + boundaries/snapshot suites).
- `go test -count=1 ./internal/envprofile/ -run 'TestDraftInstall|TestAdmitPathSource|TestDraftPath'` -> ok, exit 0.
- Legacy regression (frozen v1 untouched):
  `-run 'TestLegacyPathInstallIgnoresStoreOverlap|TestInstallPathAndList|TestReinstallSameSourceIsAnUpdate|TestFileOperandIsRefused'` -> ok, exit 0.
- `go vet ./internal/snapshot/ ./internal/envprofile/` -> clean, exit 0. `gofmt -l` on both dirs -> empty.
- Full module suite NOT run (host stalls; per wave-note policy narrow packages only; remote gate runs on handoff).

## Narrowing mutants (both killed, bytes restored with `cmp` identical)

- M1: deleted the pre-publication frozen-copy rehash in `PublishLocal`
  -> `TestPublishDetectsFrozenMutation` FAILS (`PublishLocal err = <nil>, want leading source_snapshot_changed`). Killed.
- M2: deleted the post-copy membership re-enumeration in `Capture`
  -> `TestCaptureDetectsAddedMember` FAILS (`Capture err = <nil>, want leading source_snapshot_changed`). Killed.
- No survivors. `capture.go` restored byte-identical (`cmp` clean), full snapshot package re-run green exit 0.

## Files changed (uncommitted)

- NEW internal/snapshot/capture.go (BuildInventory, Capture, PublishLocal, OpenLocal, LocalStoreDir, NormalizeSnapshotDigest + auth helpers).
- NEW internal/snapshot/capture_test.go (13 tests; 3 added this run).
- NEW internal/envprofile/draft_capture_test.go (3 Install-level tests).
- MOD internal/snapshot/boundaries.go (LocalAcquisition carries Outputs + RootEntries for post-copy re-enumeration).
- MOD internal/envprofile/envprofile.go (`admitPathSource` returns frozen snapshot dir; `stateForPath` installs it; `draftPathOutputs` incl. snapshot store).
- MOD internal/envprofile/boundaries_test.go (one-line adapt to new `admitPathSource` signature).

## Checklist

- [x] Scoped production behavior implemented per accepted draft contracts (audited + gap-closed this run)
- [x] Positive, negative and legacy regression checks run; exact revision (3dbefcd + uncommitted delta above) and evidence recorded
- [x] Code per task description and AC
- [x] Tests for new/changed behavior written and passing (3 new negative rows this run)
- [x] Lint clean (vet + gofmt)
- [x] Build/validation commands run after changes; build not broken
- [x] Outcome artifact attached (this file)
- [x] No Stop-The-Line blocker; no findings beyond the noted pre-existing uncommitted baseline

Ready for review.
