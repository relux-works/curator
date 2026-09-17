# TASK-260910-19w2aj rev7 results — resolve source closure over the pinned commit (rework 6)

Story worktree trunk: 45e397e (lock checkpoint 1a75qd) replayed onto aa46ecd.
Step 1 per 19w2aj-rework-6b.md: applied TASK-260910-19w2aj_rev6-candidate.patch
(`git apply --binary --check` exit 0, apply exit 0; status lists the 9 rev6 paths).
All rev1-rev5 fixes preserved (SkillsRoot, exact-alias recovery, legacy entries,
full Git inventory authentication, real Git selector marker, refresh rollback).
Rework-6b delta (this rev): `internal/closure/resolve.go` only, plus tests.

## Fix

`ResolveDraft` pinned every selected Git alias to its declared commit and captured
that commit's immutable tree BEFORE any expansion (new `pinGitAliases`: fetch once
per alias, `GitResolve` exactly once per alias, `snapshot.Get` the commit tree).
Baseline `Expand`, `BuildExpanded` closure, and the membership recheck all run over
`frozenGitRoots` (pinned commit trees); the caller's `GitRoots` map is never
mutated, so machine bindings still record the proving checkout and frozen
consumption still authenticates cache bytes against it. Per-selection `acquireGit`
became `acquireGitPinned` (no second resolution; subtree served from the pin).
Only selected aliases resolve; unknown aliases still surface from expansion.

## Production-entry regression tests (all through `project resolve`/`refresh` + install)

- `TestProjectResolveGitTagDiffersFromHEADThroughCLI` (cmd/curator/draft_sources_test.go):
  individual selection, tag v1 valid while HEAD removed SKILL.md. Old code refuses
  (`source_member_invalid ... SKILL.md: no such file or directory` from the checkout);
  new code locks tag v1, dry-run + real install green, marker tag v1 + commit,
  pinned reinstall green. PASS (20.19 s).
- `TestProjectResolveGitCollectionPinnedToTagThroughCLI`: collection at v1 holds
  alpha+beta while HEAD removed beta. Old code silently locks alpha alone; new code
  locks both at the tag commit, dry-run + real install + pinned reinstall green.
  PASS (19.30 s).
- `TestProjectRefreshGitBranchMembershipChangeThroughCLI`: branch-main collection
  resolves alpha; upstream gains beta; checkout HEAD provably unchanged
  (asserted equal to the pre-push commit) while refresh locks alpha+beta at the new
  commit, lock SHA moves, dry-run green, bare deleted, pinned reinstall green.
  PASS (8.60 s).
- Closure unit: `TestResolveDraftExpandsOverResolvedCommitNotCheckout`
  (individual + collection over a diverged checkout) and
  `TestResolveDraftResolvesEachGitAliasOnce` (counting resolver observes exactly
  1 call for two selections sharing one alias). PASS.

## Mutants (narrowing: `frozenGitRoots` returns the checkout roots)

- Closure: `TestResolveDraftExpandsOverResolvedCommitNotCheckout` FAILs
  (`source_member_invalid ... SKILL.md: no such file or directory`). Restored: ok.
- CLI: all three new tests FAIL under the mutant with the reviewer symptoms
  (individual refusal from `home/draft-git/...`; collection locks alpha alone;
  refresh keeps alpha alone); fix restored, `go build ./...` exit 0.

## Regression evidence (exit codes observed in-session)

- `go test ./internal/closure/ -p 1 -count=1` → ok 71.186 s, exit 0.
- `go test ./cmd/curator/ -p 1 -count=1 -run <draft batches>` → all ok, exit 0:
  local/legacy/draft-off/git-selection batch 9.904 s; tamper/missing/transitive
  batch 31.109 s; alias/legacy-configured/legacy-network/legacy-mixed/real-install
  batch 107.901 s; three new tests as above.
- `go test ./internal/install/ -p 1 -count=1 -run Draft` → ok 36.540 s, exit 0.
- `gofmt -l cmd/curator/ internal/closure/` → empty, exit 0.
- `go vet ./cmd/curator/ ./internal/closure/` → ok, exit 0.
- `golangci-lint run internal/closure/...` → 0 issues, exit 0.
- `golangci-lint run cmd/curator/...` → 0 issues, exit 0.

## Checklist

- [x] Rework 6 implemented: selections/collections expand over the resolved commit's
      frozen tree; proving repository stays separate for identity/authentication.
- [x] Previously accepted fixes preserved (rev1-rev5 + rev6 candidate).
- [x] CLI regressions for individual, collection, and branch-refresh cases, all
      driven through production `project resolve`/`refresh` + install.
- [x] Exactly-once resolution unit test; narrowing mutant fails every new test.
- [x] Narrow tests only (`-p 1`, `-run` masks); tool calls kept under 2 minutes
      except the pre-existing heavy CLI batches, which were split per instructions.
- [x] Lint clean (gofmt, vet, golangci-lint).
- [x] Workspace holds only the 9 rev6 paths; work uncommitted; story_final.

Handoff: rev7 (story_final). A `change_request_base_authority_mismatch` refusal at
handoff is handled by the orchestrator — do not retry.
