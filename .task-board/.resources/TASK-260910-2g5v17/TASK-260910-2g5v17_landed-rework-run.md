# Landed-rework run — TASK-260910-2g5v17 (P4): re-publish the exact landed tree for review

**READ THIS FIRST. This is NOT an integration run.** Your predecessor
RUN-260918-689770 (`TASK-260910-2g5v17_integration-rev2.md`) already proved
that `worktree integrate` (board_owner_separate), `worktree complete`
(code_landing_tree_mismatch) and `close-landed` (landed_tree_not_on_trunk)
all refuse revision 2 — do NOT run any of them again, do not re-verify
them, do not write another integration report. The runtime's generic
"integration assignment" text does not apply here; the orchestrator routes
this leaf through the exact-tree re-review that the `complete` refusal
itself prescribes. The ONLY board commands you run are, in this order:
`worktree obligations`, `worktree prepare-landed-review`, `add_resource`,
`worktree start-landed-rework`, `git apply` of the export's update patch in
the Story worktree, `add_resource`, `check_item` if needed, `handoff`.

You are the tracked developer run bound to the accepted revision 2 of
`TASK-260910-2g5v17`, the single (story-final) leaf of `STORY-260910-stz5f0`.
The code was landed outside the board (this repository's board lives in
`../curator`): curator-skill-registry PR #11, fast-forward push of the signed
commit `bf5cac1200cfa39dff0a6b0449072ff5b22f124d` (tree `9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`) onto `main`.
That tree is NOT the accepted candidate tree `1dfd8513e036b3e2006dd6fc620e396ad2638ffb`:
it was rebased onto the R6 landing `fb86420` (CHANGELOG entries unioned) and
the board artifact `TASK-260910-2g5v17_results.md` that the candidate carried
at the repository root was dropped (it is a task resource, not source). The
canon for a changed landed tree is an exact-tree re-review: export → rework
→ handoff → independent review → accept → complete.

Do exactly this, from the control root
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry`,
quoting every command and its full output in your outcome:
1. `task-board worktree obligations` (expect `TASK-260910-2g5v17  2  accepted  checkpoint`).
2. `task-board worktree prepare-landed-review STORY-260910-stz5f0 --cr TASK-260910-2g5v17 --revision 2 --landed-commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d --json > /tmp/2g5v17-landed-review.json`
   — read-only export. Quote the JSON (digests, `observed_landing.tree_oid`,
   the review and candidate-update patch summaries). If it refuses, quote
   the typed error and stop.
3. Attach the export as task evidence:
   `task-board m 'add_resource(TASK-260910-2g5v17, name="TASK-260910-2g5v17_landed-review-export.json", path="/tmp/2g5v17-landed-review.json", type=outcome, description="prepare-landed-review export for landed commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d")'`.
4. `task-board worktree start-landed-rework STORY-260910-stz5f0 --cr TASK-260910-2g5v17 --revision 2 --landed-commit bf5cac1200cfa39dff0a6b0449072ff5b22f124d --export /tmp/2g5v17-landed-review.json`
   — releases the leaf to to-dev with recovery intent. If interrupted, re-run
   the same command once before applying anything.
5. Apply ONLY the export's `update_patch` to the previous candidate in its
   managed Story worktree `<control-root>/.temp/STORY-260910-stz5f0/worktree`
   (`git apply` of the update patch as the export instructs; do not
   rebase or reset the branch by hand). Then verify the snapshot tree equals
   `observed_landing.tree_oid` (`9b7d33115bd3ffe44c34c5a340de72aaab06d0cb`):
   compute it with a temporary index exactly like `scripts/remote-gate.sh`
   (`GIT_INDEX_FILE=$tmp git read-tree HEAD; git add -A .; git write-tree`).
   The worktree must no longer contain `TASK-260910-2g5v17_results.md`.
6. Write a short `/tmp/2g5v17/results-rev3.md` (outside the worktree!) stating
   what changed between the accepted and the landed tree (rebase over R6,
   CHANGELOG union, stray artifact removed) and attach it as
   `TASK-260910-2g5v17_results-rev3.md` (type=outcome); tick the checklist
   if any item became unchecked; then
   `task-board handoff TASK-260910-2g5v17 --role developer`. The hosted gate
   runs once at handoff. No other changes.
