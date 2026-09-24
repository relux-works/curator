# TASK-260916-2ok97n (R5) — rework 6 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION. HIGHEST PRIORITY.

Revision 6 is green everywhere but was CHANGES_REQUESTED for F1 (`TASK-260916-2ok97n_review-verdict-rev6*`): the candidate contains
downloaded CI artifacts and task documents — 24 root paths, ~90k lines: everything under `test/` and `ledger/` (a `gh run download`
of a Windows gate artifact; the orchestrator's earlier brief told you to download "from your worktree" — that was wrong) and the root
files `TASK-260916-2ok97n_results.md`, `TASK-260916-2ok97n_handoff-blocker.md`.
1. Delete exactly those paths from the worktree (they are untracked additions; do not touch any other file). Keep every product/test/
   docs/CHANGELOG/.github/ci change of revision 6 byte-identical.
2. Make sure the board copies of the results/blocker documents are current (`task-board resource update … --type outcome`).
3. Prove: `git status --short` and the diff stat vs the Story base contain only product, test, docs, CHANGELOG and `.github/ci` paths;
   the non-removed paths are byte-identical to revision 6 (per-file hash list in results).
4. Append "Revision 7 (artifact cleanup)" to the results resource, `resource update` it, then `task-board handoff
   TASK-260916-2ok97n --role developer`; stay in the turn while the gate runs. Evidence downloads, if any, go to $TMPDIR only.
Do NOT run refresh-candidate. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
