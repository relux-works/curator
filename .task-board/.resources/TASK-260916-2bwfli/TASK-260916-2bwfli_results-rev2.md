# TASK-260916-2bwfli results rev2 — Windows gate repair (rework of rev1)

Producer: developer (rework run). Shell: bash in the Story worktree
(`task-board/story/STORY-260910-1bhj0g`). All exit codes below are real.
Rev1 was functionally complete on unix; the remote gate failed ONLY on
windows-latest (run 35150651144): one genuine test failure plus
unrecognised skip reasons. This rev fixes both; no production-code
change beyond rev1.

## Root causes (from the gate evidence artifact)

1. `go test exit=1`: `cmd/curator :: TestDraftTransportWindowsRefusal`
   expected the typed `transport_resolution_unsupported_platform` on
   CLI output, but `buildrepo.RunPipeline` masks EVERY acquisition
   failure as `build_repository_source_unavailable` (pipeline.go:195).
   Observed stderr: `skill-a.https-cmd:
   build_repository_source_unavailable: exact external source is
   unavailable`. The typed code is unobservable above the pipeline.
2. Platform-case gate: 12 skips with reason `stand-in git is a POSIX
   shell wrapper`, which matches no row of
   `.github/ci/skip-classes.tsv` (fatal by design).

## Delta over rev1 (4 files, tests + 1 docs sentence)

- `cmd/curator/draft_transport_test.go`: 3 skip sites now use the
  recognised 5nrmtt vocabulary `test transport wrapper is POSIX-only;
  production code is platform-neutral` (skip-classes.tsv
  platform-control/allow). Deleted `TestDraftTransportWindowsRefusal`
  (unobservable assertion; replaced by a pointer comment).
- `internal/install/drafttransport_test.go`: same skip fix in
  `draftGitTool`; new Windows-only
  `TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess` calls the
  production caller `acquireDraftNetwork` (switch on, valid policy,
  unrunnable tool a la the executor's own probe) and asserts the typed
  `CodeTransportResolutionUnsupportedPlatform`. Off-Windows skip reason
  `windows refusal probe is exercised on Windows` matches
  `(is|are) exercised (on|by)` (platform-control/allow).
- `docs/draft-transport-resolution.md`: caller section notes the
  pipeline masking (typed Windows code visible only below it).
- No `.github/ci` change: vocabulary reuse keeps the skip classes closed.

## Evidence (narrow, `-p 1 -count=1`)

| command | exit |
|---|---|
| `gofmt -l` on all 5 touched Go files | 0, no output |
| `go vet ./internal/install/` | 0 |
| `go vet ./cmd/curator/` | 0 |
| `golangci-lint run ./cmd/curator/ ./internal/install/` | 0, 0 issues |
| `go test ./internal/install/ -run 'TestDraftTransportEnabled\|...\|TestAcquireDraftNetworkSelection\|TestAcquireDraftNetworkWindowsRefusesBeforeAnyProcess\|TestDefaultAcquireFetchesTheDeclaredURL'` | 0, ok 13.8s |
| `go test ./cmd/curator/ -run 'TestDraftTransportLegacyGolden\|TestDraftTransportResolvedMatrix\|TestProductionBinaryDispatchesSSHWrapper'` | 0, ok 51.0s, all PASS |
| mutant: `if !deps.DraftTransportResolution` -> `if false`, golden | 1 FAIL (ssh line shows the manager wrapper) — killed |
| mutant reverted (`grep -c MUTANT` = 0), golden re-run | 0, ok 23.0s |
| `GOOS=windows go vet ./cmd/curator/ ./internal/install/` | 0 (both test binaries compile; execution impossible on darwin — expected `exec format error`) |
| skip-reason grep vs skip-classes.tsv regexes | both match `allow` rows; no `t.Skip` keeps the old text (1 comment mention only) |

Executor untouched: `git diff --stat` and status over
`internal/buildrepo` + `internal/config` are empty. Workspace holds
exactly the 7 intended files (3 modified, 4 new); golden md5
af83b55eb42e926b1b9b718701ebc500, unchanged by this rev. Windows
run-verdict comes from the remote gate on handoff (this host cannot
execute Windows binaries); compile-verified here, lane refusal itself
covered by the executor's own Windows probe (5nrmtt, untouched).
