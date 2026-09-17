# TASK-260910-19w2aj results — rev5 (rework 4: SkillsRoot + alias selection)

Two findings from verdict rev4, both fixed at the production entry and
covered by CLI tests that drive `project resolve` and install through the
CLI with a temp config. Rev2/rev3 fixes kept unchanged. Workspace holds
only the 1a75qd checkpoint + this leaf's delta (story_final); no commits
on the Story branch.

## Changes (uncommitted in Story worktree task-board/story/STORY-260910-3vxe3y)

1. P1 — `cmd/curator/project_resolve.go` (`resolveDraftPlan`): pass the
   configured dependency repository root into the draft resolve config
   (`SkillsRoot: cfg.SkillsRoot`). The transitive/legacy lane
   (`internal/closure/closure.go:359`) joins it with the provider name;
   with the empty root it resolved relative to the process CWD instead
   of `<configured skills root>/provider`.
2. P2 — `internal/install/draftsources.go` (`draftFrozenInput`): recover
   the declaring Git alias of a locked root member from the lock
   member's selection index (`skillfile-sources §3 binds each root
   member to its declaring selection`) via new `declaringGitSource`
   instead of the first alias sharing the canonical repository
   identity. Bindings lookup and declared ref (kind/value/endpoint)
   both follow the indexed alias; anything else fails closed with
   `source_member_invalid`. Removed the now-unused `sort` import and
   updated the function doc comment.

## New CLI tests (`cmd/curator/draft_sources_test.go`, lock always created
through `project resolve`, never by calling the helper under test)

- `TestProjectResolveTransitiveProviderUsesConfiguredRootThroughCLI`:
  provider checkout only under the temp config's skills root, declared
  requirement URL unroutable (`https://invalid.invalid/...`) so any
  clone/fetch fails; resolve must succeed (zero acquisitions), lock
  must carry provider with no root selection index, `install
  --dry-run` green.
- `TestProjectResolveTransitiveProviderIgnoresProcessCWDThroughCLI`:
  provider checkout only under the process CWD (`t.Chdir`);
  configured root has none; resolve must fail with `failed to clone`
  (falls through to acquisition, never the CWD copy) and publish no lock.
- `TestProjectResolveGitAliasSelectionThroughCLI`: aliases a→v1 and
  s→v2 on the same commit, selection `from: s`; dry-run stdout must
  contain `review tag v2 `, real install marker must read tag/v2/commit.

## Evidence (Story worktree, `set -o pipefail` shells; real exit codes)

- `gofmt -l cmd/curator internal/closure internal/install` → clean, exit 0
- `go vet ./cmd/curator/ ./internal/install/ ./internal/closure/` → exit 0
- `golangci-lint run cmd/curator/... internal/closure/... internal/install/...`
  → `0 issues`, exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestProjectResolveTransitiveProvider(UsesConfiguredRoot|IgnoresProcessCWD)ThroughCLI'
  -v ./cmd/curator/` → both PASS (0.96s/0.34s test; 147s wall cold
  test-binary build), exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestProjectResolveGitAliasSelectionThroughCLI' -v ./cmd/curator/`
  → PASS (6.08s test; 7s wall warm cache), exit 0
- `go test -p 1 -count=1 -timeout 9m -run 'TestProjectResolve' -v
  ./cmd/curator/` → all 10 PASS (incl. rev2/rev3 rows: local lock,
  legacy untouched, draft-off, git selection, runtime tamper, missing
  member, real install), exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestResolveDraft|TestOpenDraftFrozen|TestRefresh|TestResolveDraftGitSubtree|TestProjectResolve|TestDraftTransitive|TestLoadDraft|TestDraftFrozen'
  ./internal/closure/` → ok, exit 0
- `go test -p 1 -count=1 -timeout 9m -run
  'TestDraftInstall|TestLegacyInstallUntouched|TestDraftSourcesSwitch'
  ./internal/install/` → ok, exit 0

## Mutants (narrowing, both killed; fixes restored byte-identical, md5 verified)

- Drop `SkillsRoot` from `resolveDraftPlan`: positive test FAILS
  (`failed to clone provider ... Could not resolve host:
  invalid.invalid`, proving the green run used the configured root
  with zero acquisitions); CWD test FAILS (resolve exits 0 via the CWD
  checkout, proving the fixed code never consults CWD).
- First-alias recovery (ignore selection index): alias test FAILS with
  `app: review tag v1 1649d0b ... (planned)`, proving the report and
  marker bind the declaring selection `s` → v2.

## Checklist

- [x] Production resolve carries `cfg.SkillsRoot` into the draft closure
- [x] Frozen Git identity recovered from lock selection index, not first alias
- [x] CLI tests drive `project resolve` + install with temp config only
- [x] Mutants killed for both findings; survivors: none
- [x] Lint clean (gofmt, vet, golangci-lint)
- [x] Narrow tests green with real exit codes (packages touched only)
- [x] Workspace holds only 1a75qd checkpoint + leaf delta; no commits
- [x] Rev2/Rev3 fixes untouched (CLI entry, Git subtree, full-inventory
      auth, transactional refresh, POSIX skip reason verbatim)
