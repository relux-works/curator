# Review brief — TASK-260918-1hl5f8: landing review of the rebased E6 spec tree

Round 2 (`TASK-260916-3l60rn_review-verdict-rev2.md`) ACCEPTED the S5
candidate (`TASK-260916-3l60rn_spec-patch_rev2.patch`, base curator-spec
`e8b53a0`). Since then curator-spec `main` moved to `4a2fa3e` (E3 #69 landed) on
overlapping paths, so the orchestrator rebased the accepted
candidate for landing. The board Change Request of the S5 task stays
accepted (empty curator delta); this round reviews the EXACT tree that will
land:

- delivery worktree (read-only for you):
  `/Users/administrator/Developer/ReluxWorks/.worktrees/curator-spec-e6` at
  commit `1ca4b3d` = head of https://github.com/relux-works/curator-spec/pull/70
  (branch `e6-path-kind-admission`, parent `4a2fa3e`);
- the union diff is attached as `e6-union-vs-4a2fa3e.patch`; the accepted
  candidate patch (vs `684c9f1`) is attached as
  `TASK-260916-3l60rn_spec-patch_rev2.patch`.

## What to verify
1. **Merge fidelity.** For every file of the union diff, either its hunks are
   byte-identical (modulo context) to the accepted candidate's, or the file
   is one of the merged sources: `protocol/environments.md` (exactly TWO
   hand-composed hunks — the §12 status paragraph, where E6's rewritten
   store-trust row sentence ("… named store entries, and — for a `path` root
   or overlay — the source directory, naming the failing check and the path
   or, for an enclosing failure, the boundary …") is followed by E3's
   codex-seed rows in one list ending "Both commands follow"; and the §13
   conformance surfaces, where the E5, S5, E3 and E6 blocks are joined as
   "…non-conforming; the section 7.4 codex-seed cases …; and the section 2.2
   and section 4 path-kind admission cases …"), `tools/validate.py` (both
   `validate_environments_codex_seed_vectors` and
   `validate_environments_path_kind_admission_vectors` registered in
   `main()`), or a regenerated file (`conformance/v1/manifest.json`,
   `release/1.0.0-rc.9.json`). For each merged hunk confirm the result is the
   UNION of the landed content (compare with `4a2fa3e`) and the accepted E6
   content (compare with the rev2 patch): nothing landed was dropped,
   nothing of E6 was dropped, no third content was invented, no sentence
   contradicts another. Quote file:line.
2. **Regeneration exactness.** `make regenerate-check` in a disposable byte
   copy with a temporary git baseline exits 0; the manifest/pins equal a
   fresh regeneration.
3. **Gates.** `make validate` from a disposable copy with the repository venv
   (`PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`),
   exit codes quoted.
4. **Nothing else.** `git diff 4a2fa3e..1ca4b3d --stat` lists only the files
   of the accepted candidate plus the regenerated ones; the delivery worktree
   is read-only for you — no edits, no files left there.

## Verdict
Record `TASK-260918-1hl5f8_review-verdict.md` (task outcome) with the
per-hunk table (quotes, file:line, verdict per hunk), transcripts and
findings; then set this task `done` with
`task-board m 'set_status(TASK-260918-1hl5f8, status=done)'` when the tree
may land, or `to-dev` with the concrete corrections when it may not (the
orchestrator repairs the delivery tree; the accepted E6 record is not
touched). Never edit the delivery worktree.
