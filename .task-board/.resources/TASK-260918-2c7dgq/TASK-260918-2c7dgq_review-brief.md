# Review brief — TASK-260918-2c7dgq: landing review of the rebased E3 spec tree

Round 2 (`TASK-260916-2rnkei_review-verdict-rev2.md`) ACCEPTED the S5
candidate (`TASK-260916-2rnkei_spec-patch_rev2.patch`, base curator-spec
`684c9f1`). Since then curator-spec `main` moved to `e8b53a0` (R3/P2 #66, E5 #67 and
S5 #68 landed) on overlapping paths, so the orchestrator rebased the accepted
candidate for landing. The board Change Request of the S5 task stays
accepted (empty curator delta); this round reviews the EXACT tree that will
land:

- delivery worktree (read-only for you):
  `/Users/administrator/Developer/ReluxWorks/.worktrees/curator-spec-e3` at
  commit `4a2fa3e` = head of https://github.com/relux-works/curator-spec/pull/69
  (branch `e3-codex-seed`, parent `e8b53a0`);
- the union diff is attached as `e3-union-vs-e8b53a0.patch`; the accepted
  candidate patch (vs `684c9f1`) is attached as
  `TASK-260916-2rnkei_spec-patch_rev2.patch`.

## What to verify
1. **Merge fidelity.** For every file of the union diff, either its hunks are
   byte-identical (modulo context) to the accepted candidate's, or the file
   is one of the merged sources: `protocol/environments.md` (exactly TWO
   hand-composed hunks — the §12 status paragraph, where the S5 store-trust
   row sentence is followed by E3's codex-seed rows in one list ("value, the
   store-trust row … `environment_store_untrusted`, the active codex-seed
   revision … revision-`B` manager). Both commands follow"); and the §13
   conformance surfaces, where the E5 write-nofollow block, the S5
   store-boundary block and E3's codex-seed block are joined as "…; the
   section 4 …; and the section 7.4 codex-seed cases …"), `tools/validate.py`
   (both the S5 and the E3 constant blocks kept, and both
   `validate_environments_store_boundary_vectors` and
   `validate_environments_codex_seed_vectors` registered in `main()`),
   `CHANGELOG.md` (both entries, additive), or a regenerated file
   (`conformance/v1/manifest.json`, `conformance/v1/schema-cases/index.json`,
   `release/1.0.0-rc.9.json`, the marker schema cases). For each merged hunk
   confirm the result is the UNION of the landed content (compare with
   `e8b53a0`) and the accepted E3 content (compare with the rev2 patch):
   nothing landed was dropped, nothing of E3 was dropped, no third content
   was invented, no sentence contradicts another. Quote file:line.
2. **Regeneration exactness.** `make regenerate-check` in a disposable byte
   copy with a temporary git baseline exits 0; the manifest/pins equal a
   fresh regeneration.
3. **Gates.** `make validate` from a disposable copy with the repository venv
   (`PATH=/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/venv/bin:$PATH`),
   exit codes quoted.
4. **Nothing else.** `git diff e8b53a0..4a2fa3e --stat` lists only the files
   of the accepted candidate plus the regenerated ones; the delivery worktree
   is read-only for you — no edits, no files left there.

## Verdict
Record `TASK-260918-2c7dgq_review-verdict.md` (task outcome) with the
per-hunk table (quotes, file:line, verdict per hunk), transcripts and
findings; then set this task `done` with
`task-board m 'set_status(TASK-260918-2c7dgq, status=done)'` when the tree
may land, or `to-dev` with the concrete corrections when it may not (the
orchestrator repairs the delivery tree; the accepted E3 record is not
touched). Never edit the delivery worktree.
