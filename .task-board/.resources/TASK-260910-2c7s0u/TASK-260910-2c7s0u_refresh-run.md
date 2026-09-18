# Refresh + handoff run — TASK-260910-2c7s0u (R4 docs, story-final leaf of STORY-260910-2xe3n2)

Your predecessor RUN-260918-b6238d wrote the R4 deployment documentation
(uncommitted in the Story worktree: `CHANGELOG.md`, `README.md`, `SECURITY.md`)
and its handoff failed at Change Request construction with
`stale-anchor: change_request_base_authority_mismatch`: this is the story's
LAST open leaf, so its Change Request is built against the FRESH trunk
authority, and the Story branch (checkpoints `d7f424c` R5, `bd27d32` R8,
`5f3b028` R7 on base `c7ef32c`) does not descend from the current
curator-skill-registry `main` (`bf5cac1`: R6 `fb86420` and P4 `bf5cac1`
landed meanwhile). The sanctioned recovery is `task-board worktree
refresh-candidate`, which replays the checkpoints onto the fresh trunk in an
isolated workspace with signing, preserving your working candidate exactly.
Rules: `remediation-registry-producer-rules.md` (attached). Role: doc-writer.

Do exactly this, from the control root
`/Users/administrator/Developer/ReluxWorks/curator/curator-skill-registry`,
quoting every command and its output in `TASK-260910-2c7s0u_results.md`
(update the existing results with a "Refresh" section; write it OUTSIDE the
worktree and re-attach it):
1. `task-board worktree obligations`; `git -C <worktree> status --short --untracked-files=all`
   — expect only the three documentation files modified, nothing else.
2. Combine your working candidate with the incoming trunk content FIRST
   (refresh-candidate requires it): the R6 and P4 landings changed
   `CHANGELOG.md` (two new Unreleased entries at the top), `README.md` and
   `SECURITY.md` (passphrase-key and import high-water sections). Read
   `git -C <control-root> diff c7ef32c bf5cac1 -- CHANGELOG.md README.md SECURITY.md`
   and make sure your R4 text does not contradict or duplicate those
   sections (cross-reference where useful; do not restate them).
3. `task-board worktree refresh-candidate TASK-260910-2c7s0u`. The replay
   of the three checkpoints will stop on regular-file conflicts (a dry
   rebase shows `CHANGELOG.md` conflicts on the first checkpoint: the R5/R8/R7
   entries and the R6/P4 entries were added at the same place). It retains
   a replay worktree with `resolution-template.json` in its parent directory:
   prepare the replacement content — the UNION of both sides' CHANGELOG
   entries, newest landing first as on `main` (P4, R6), then the checkpoint's
   entry — bound to `REBASE_HEAD`, the unmerged path and the SHA-256 of the
   replacement bytes exactly as the template asks, and re-run
   `task-board worktree refresh-candidate TASK-260910-2c7s0u --replay-resolutions <file>`;
   repeat per checkpoint if the tool stops again. Never hand-commit the
   retained replay worktree, never `git rebase` the Story branch yourself,
   never edit workspace registry records.
4. After the replay: `git -C <worktree> log --oneline -5` (three replayed
   checkpoints on top of `bf5cac1`), your three documentation files still
   modified and uncommitted; run `python -m pytest -q` (venv outside the tree,
   `CURATOR_CONFORMANCE_ROOT` at the `47c3c8c` detached root) — exit 0 —
   and `python -m mypy`.
5. Tick the checklist and `task-board handoff TASK-260910-2c7s0u --role doc-writer`
   IMMEDIATELY after (the story-final Change Request is constructed against
   the trunk at that moment). If it refuses with `stale-anchor` again
   (trunk moved during the run), quote it, re-run step 3 (the replay starts
   from the unchanged original checkpoints) and hand off again — at most
   twice, then stop and report.
