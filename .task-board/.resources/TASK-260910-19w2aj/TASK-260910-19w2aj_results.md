# TASK-260910-19w2aj results — rev3 (rework 2: Windows platform-case gate)

Scope: one-line fix per `19w2aj-rework-2.md`. Nothing else changed from rev2
(the four rev1 P1 findings stay fixed as reviewed).

## Change

`cmd/curator/draft_sources_test.go:196` — `TestProjectResolveGitSelectionThroughCLI`
Windows skip reason changed verbatim to the declared platform-control class:

- before: `t.Skip("fake git wrapper is POSIX-only; production code is platform-neutral")`
- after: `t.Skip("test transport wrapper is POSIX-only")`
- declared class: `.github/ci/skip-classes.tsv:60`
  `platform-control / test transport wrapper is POSIX-only / allow`
  ("the fixture uses a POSIX shell wrapper; native admission remains covered on Windows")
- still gated on `runtime.GOOS == "windows"` only; no skip class added.
- Chose the verbatim-reason option over the Go-test-binary fake git: the
  re-exec wrapper would add a new test-binary build on an already
  build-stalled host for no product-code gain (production code is
  platform-neutral; the POSIX part is the shell fixture only).

## Evidence (all in Story worktree `task-board/story/STORY-260910-3vxe3y`, uncommitted)

- `gofmt -l cmd/curator internal/closure internal/install` → clean, exit 0
- `go vet ./cmd/curator/ ./internal/closure/ ./internal/install/` → exit 0
- `GOOS=windows go vet ./cmd/curator/` → exit 0 (skip file cross-compiles)
- `golangci-lint run cmd/curator/... internal/closure/... internal/install/...`
  → `0 issues`, exit 0
- `go test -p 1 -count=1 -run 'TestProjectResolveGitSelectionThroughCLI' ./cmd/curator/`
  → PASS (5.54s test; ~130s cold test-binary build on this host), exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestProjectResolveLocalCreatesLockThroughCLI|TestProjectResolveLegacyUntouched|TestProjectResolveDraftOffRefuses|TestProjectResolveGitSelectionThroughCLI'
  -v ./cmd/curator/` → all 4 PASS, exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestResolveDraft|TestOpenDraftFrozen|TestRefresh|TestResolveDraftGitSubtree|TestProjectResolve'
  ./internal/closure/` → ok, exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestDraftInstall|TestLegacyInstallUntouched|TestDraftSourcesSwitch'
  ./internal/install/` → ok, exit 0
- Note: one earlier cold-cache run hit go test's default 600s timeout while
  building the `cmd/curator` test binary (host stalls on new test binaries);
  rerun with warm cache passes. Not a product-code hang (test itself ~3–5s).

## Checklist

- [x] Skip reason matches `.github/ci/skip-classes.tsv:60` verbatim
- [x] Gated on `GOOS == windows` only; no new skip class
- [x] Lint clean (gofmt, vet incl. GOOS=windows, golangci-lint)
- [x] Narrow tests green with real exit codes
- [x] Workspace holds only the 1a75qd checkpoint + this leaf's delta (story_final)
- [x] No commits on the Story branch (uncommitted worktree for handoff snapshot)
