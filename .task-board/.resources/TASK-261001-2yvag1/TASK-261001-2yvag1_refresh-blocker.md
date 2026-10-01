# TASK-261001-2yvag1 — curator-muse-environment


## Revision 5 (refresh) — blocked before refresh

Run: RUN-261001-2759d9. Requested trunk: bd126a9a.

The required command `task-board worktree refresh-candidate TASK-261001-2yvag1` was executed directly and returned exit 1:

```text
candidate refresh requires a rework revision; TASK-261001-2yvag1 is ready
```

`task-board worktree status STORY-261001-1xuwlu --json` returned exit 0 and confirms revision 4 is still `ready`, base c803afd7e9ed9fe016f85be10fc366cd6dccc0e1, candidate tree 71308ba696e14ea5fba40fed5b65fa298f3effdb, with this run holding the Story lease. The failed reviewer preparation did not route this revision into rework. The earlier reviewer refusal names the supported remedy: `task-board worktree converge STORY-261001-1xuwlu --reason "trunk advanced on changed candidate paths"`. The project-management recovery reference requires that command from the control root outside a producer run; a tracked run is refused. `invalidate-acceptance` also requires an operator and an integrating element, so it does not apply to this ready revision.

Per-file resolution summary: no resolutions or repository changes were made. `.github/ci/conformance-case-counts.tsv`, `cmd/curator/env.go`, `internal/envprofile/managed.go`, and `internal/envprofile/status.go` retain revision 4 contents. Neither side's behavior has been combined yet. Counts were not recomputed against a refreshed candidate.

Not run in this attempt: `go test ./cmd/curator ./internal/envprofile -count=1`, the 16 Muse link-state rows, the 36 fragment-v3 rows, the six env-status/guard regression tests, TestPosture* and subcommand --help rows, build, lint, and mutants. The refresh failed before a revision 5 source tree existed; running these against the old base would not validate the required combined tree. Prior revision 4 hosted green evidence (run 36833591988) is historical evidence only, not revision 5 evidence.

Constraint: board lifecycle and operator-owned workspace convergence, not a product or source-code defect. Manual Git rebase/merge/commit, direct board edits, self-issued reviewer rejection, or withdrawal would bypass the managed lifecycle and are not viable repairs.

Recommended external action: after this run releases its Story lease, the orchestrator runs the named `converge` command from the control checkout, resolves any refusal through its supported instructions, and launches a developer rework run once revision 4 is stale/rework eligible. Then refresh, preserve both sides, run the requested checks with real exits, recompute exact digest counts, attach evidence, and hand off revision 5. No human product decision is needed.

No developer handoff was attempted because the required refresh and validation have not occurred. No launcher, askpass, CHANGELOG, or LOGBOOK changes were made. The separately tracked askpass EPIPE issue remains out of scope. Logbook entry text: revision 5 refresh was refused because revision 4 remained ready after reviewer preparation failed; operator convergence is required before developer rework.
