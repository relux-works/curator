# Review brief — TASK-260917-e6nwdj: landing review of the rebased S5 spec tree

Round 2 (`TASK-260910-39fzpq_review-verdict-rev2.md`) ACCEPTED the S5
candidate (`TASK-260910-39fzpq_spec-patch_rev2.patch`, base curator-spec
`684c9f1`). Since then curator-spec `main` moved to `9912db7` (R3/P2 #66 and
E5 #67 landed) on overlapping paths, so the orchestrator rebased the accepted
candidate for landing. The board Change Request of the S5 task stays
accepted (empty curator delta); this round reviews the EXACT tree that will
land:

- delivery worktree (read-only for you):
  `/Users/administrator/Developer/ReluxWorks/.worktrees/curator-spec-s5` at
  commit `e8b53a0` = head of https://github.com/relux-works/curator-spec/pull/68
  (branch `s5-store-boundary`, parent `9912db7`);
- the union diff is attached as `s5-union-vs-9912db7.patch`; the accepted
  candidate patch (vs `684c9f1`) is attached as
  `TASK-260910-39fzpq_spec-patch_rev2.patch`.

## What to verify
1. **Merge fidelity.** For every file of the union diff, either its hunks are
   byte-identical (modulo context) to the accepted candidate's, or the file
   is one of the merged sources: `protocol/environments.md` (exactly FOUR
   hand-composed hunks where S5 and E5 meet — the §10.1 link-target currency
   sentence: S5's "necessary but no longer sufficient … verified, not
   assumed (section 4)" + E5's "`lstat`-class semantics (section 8.3.1)"; the
   §10.1 repair paragraph: E5's "Repair writes are section 8.3.1 writes …"
   sentence followed by S5's failure-class paragraph; the §12 non-current
   list: both `store-untrusted (environment_store_untrusted)` and
   `link-blocked (section 8.3.1)`; the §13 conformance surfaces: E5's
   write-nofollow block and S5's store-boundary block joined with ";
   and"), `CHANGELOG.md` (both entries, additive), `tools/validate.py` and
   `tools/test_validate.py` (line-identical unions of the two validator
   families), or a regenerated file (`conformance/v1/manifest.json`,
   `release/1.0.0-rc.9.json`). For each merged hunk confirm the result is the
   UNION of the landed E5 content (compare with `9912db7`) and the accepted
   S5 content (compare with the rev2 patch): nothing landed was dropped,
   nothing of S5 was dropped, no third content was invented, no sentence
   contradicts another (e.g. the currency sentence must not both claim
   "sufficient" and "no longer sufficient"). Quote file:line.
2. **Regeneration exactness.** `make regenerate-check` in a disposable byte
   copy with a temporary git baseline exits 0; the manifest/pins equal a
   fresh regeneration.
3. **Gates.** `make validate` from a disposable copy with the repository venv
   (`PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`),
   exit codes quoted.
4. **Nothing else.** `git diff 9912db7..e8b53a0 --stat` lists only the files
   of the accepted candidate plus the regenerated ones; the delivery worktree
   is read-only for you — no edits, no files left there.

## Verdict
Record `TASK-260917-e6nwdj_review-verdict.md` (task outcome) with the
per-hunk table (quotes, file:line, verdict per hunk), transcripts and
findings; then set this task `done` with
`task-board m 'set_status(TASK-260917-e6nwdj, status=done)'` when the tree
may land, or `to-dev` with the concrete corrections when it may not (the
orchestrator repairs the delivery tree; the accepted S5 record is not
touched). Never edit the delivery worktree.
