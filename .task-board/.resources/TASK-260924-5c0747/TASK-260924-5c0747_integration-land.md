# TASK-260924-5c0747 integration preconditions — CONFIRMED (bound developer run, 2026-09-26)

Revision 3 of CR-TASK-260924-5c0747 is ACCEPTED and ready for the runner's
bound landing. No file was changed and no board status was written by this run.

## Preconditions (all verified read-only in STORY-260924-2go2bz worktree)

1. Board status: `integrating` (task-board q get → `{"id":"TASK-260924-5c0747","status":"integrating"}`).
2. Acceptance: `TASK-260924-5c0747_review-verdict-rev3.md` records rev3 ACCEPTED
   by reviewer claude-opus-5-5 (low), candidate base f02ba39e, tree 4c96bc50, 14 paths.
3. Gate green: `TASK-260924-5c0747_change-request_rev3-validation.log` ends
   `[exit 0]` — remote gate run 36260210792 success on all jobs
   (Test ubuntu/macos/windows, Race, Interop conformance gate, Gate self-tests;
   rose-air skipped; required=1 green=1).
4. Trunk fresh and unmoved: worktree HEAD = f02ba39e, `git fetch origin main`
   this run → origin/main = f02ba39e4c6f6d12af950c5138e8ffd7bbe8bb95. Matches the
   reviewed base; no trunk movement to reconcile.
5. Delta identity: worktree path set (13 modified + 1 untracked manifest.json =
   14 paths) is IDENTICAL to the `diff --git` path list in
   `TASK-260924-5c0747_change-request_rev3.patch` (sorted `diff` → no output).
   All paths are product/test/docs/CHANGELOG/.github only — no stray files, no
   `.task-board` or CI-artifact contamination in the candidate tree.
6. No commit past checkpoint: HEAD == f02ba39e (changes uncommitted, nothing
   staged as a commit); no `change_request_candidate_committed_past_checkpoint`
   risk. (One stash entry exists but belongs to STORY-260906-1a2i5a — untouched.)
7. Pin spot-check: `SPEC_PIN: 23435129ebc4c29e5b7f75ec72a0aa0cd3f16065` in
   `.github/workflows/ci.yml`; `SKILLFILE_SOURCES_PIN` = same commit. Sole
   `dcc7f015` occurrence in the tree is CHANGELOG.md:445 history prose
   describing the superseded interim pin — no active reference remains.

## For the orchestrator (from the accepted review, unchanged)

- After the landing commits: `git fetch origin main && git tag -s v0.15.0-rc.2 -m "v0.15.0-rc.2" origin/main && git push origin refs/tags/v0.15.0-rc.2`
- Watch: `gh run list --workflow release.yml --ref v0.15.0-rc.2` + `gh run watch --exit-status`

## Note on the integrate-land instruction

`5c0747-integrate-land.md` orders running `task-board worktree integrate` here;
the binding Integration Assignment for this run supersedes it and forbids
executing `worktree integrate`/`checkpoint` in-turn — the runner performs the
bound landing synchronously after this evidence is attached. This run therefore
performs precondition confirmation only and ends with the board still at
`integrating`, no handoff call, no status write.
