# THE ONLY CURRENT INSTRUCTION — TASK-261001-2yvag1 base refresh → revision 5 (orchestrator brief, binding)

Revision 4 is green (hosted run 36833591988) but has not been reviewed yet. The reviewer spawn was refused with revision_base_superseded: curator main moved on paths this candidate also changes:
- .github/ci/conformance-case-counts.tsv
- cmd/curator/env.go
- internal/envprofile/managed.go
- internal/envprofile/status.go

Landed since base c803afd7:
- 3b3oyi (second-operator first-run UX: posture warning bound to the parent pid, subcommand --help restored);
- 20ao7p (external-repo black-box);
- 1uepyd (clean-git admission).

Current main: bd126a9a.

1. Refresh onto current trunk: `task-board worktree refresh-candidate TASK-261001-2yvag1`.
   - On a conflict, follow its `--replay-resolutions` instructions. Never hand-commit the replay worktree.
   - Keep BOTH sides' behaviour. 3b3oyi's posture/--help semantics in env.go/status.go stay intact.
   - conformance-case-counts.tsv must carry both sides' rows with exact counts. Recompute; do not guess.
2. Re-run with real exit codes:
   - `go test ./cmd/curator ./internal/envprofile -count=1`
   - the Muse rows (16 link-state, 36 fragment-v3)
   - the six formerly failing env-status/guard tests
   - the 3b3oyi rows (TestPosture*, the subcommand --help rows)
3. In the results, add a "Revision 5 (refresh)" section with the per-file resolution summary. Then run `task-board handoff TASK-261001-2yvag1 --role developer` and END YOUR TURN.

No product change beyond the conflict resolution. Never spell any employer name. Do not touch askpass code (BUG-261001-2772iz).

## Update (orchestrator, binding): convergence done
The orchestrator ran `task-board worktree converge STORY-261001-1xuwlu`.
- The workspace base is now bd126a9a.
- The 28 uncommitted candidate paths were carried over, and rev 4 is now stale, so rework is eligible.
- The four intersecting paths need your review: .github/ci/conformance-case-counts.tsv, cmd/curator/env.go, internal/envprofile/managed.go, internal/envprofile/status.go.

Do this:
1. Check `git status` and `git diff` in the Story worktree. Make sure the carried delta on those four files correctly combines with trunk: no conflict markers, no lost 3b3oyi behaviour, counts recomputed exactly.
2. If `refresh-candidate` is still needed or offered, use it. Otherwise resolve in the worktree.
3. Run the checks from item 2 above with real exit codes, then hand off revision 5.
