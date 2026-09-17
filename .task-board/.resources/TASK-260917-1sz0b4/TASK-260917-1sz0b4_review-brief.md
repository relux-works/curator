# Review brief — TASK-260917-1sz0b4: landing review of the rebased S4 spec tree

Round 3 (`TASK-260910-2ohnjo_review-verdict-rev3.md`) ACCEPTED the S4 candidate
(`TASK-260910-2ohnjo_spec-patch_rev3.patch`, base curator-spec `07e2b41`).
Since then curator-spec `main` moved to `0da4020` (S6 #60, E2 #61, E4 #62 landed)
on overlapping paths, so the orchestrator rebased the accepted candidate for
landing. The board Change Request stays accepted (empty curator delta); this
round reviews the EXACT tree that will land:

- delivery worktree (read-only for you):
  `/Users/administrator/Developer/ReluxWorks/.worktrees/curator-spec-s4-passthrough`
  at commit `23dafa798fa80fc2591ddb287c1c6345e2715b3b` = head of
  https://github.com/relux-works/curator-spec/pull/63 (branch
  `spec/s4-mcp-passthrough-bounds`, parent `0da4020`).

## What to verify
1. **Merge fidelity.** `git diff 0da4020..23dafa7` vs the accepted patch: for
   every file, either the hunks are byte-identical / context-only, or the file
   is one of the merged sources — `CHANGELOG.md` (both entries), `protocol/environments.md`
   (§12 status paragraph, diagnostic list, §13 surfaces), `tools/generate-vectors/manager_config.go`
   (defaults map with S4's `[]` default plus the E2/E4 keys; both case lists) —
   or a regenerated file (manifest, `vectors/manager-config-v2.json`, release
   pins). For each merged source confirm the result is the UNION of the landed
   S6/E2/E4 content (compare with `0da4020`) and the accepted S4 content
   (compare with the rev3 patch): nothing landed was dropped, nothing of E4 was dropped, no
   third content was invented. Quote file:line.
2. **Regeneration exactness.** `make regenerate-check` in a disposable byte
   copy with a temporary git baseline (as in earlier rounds) exits 0; the
   regenerated `manager-config-v2` vector and cases carry the S4 `[]` default
   together with the E2/E4 knobs consistently.
3. **Validation.** `make validate` (repo venv on PATH, `set -o pipefail`) —
   quote the three gate outputs and exit code. Re-run the round-1/2 mutants (unlisted name passed under `s4-enforce`;
   `INVALID\n` surfacing bytes; `sourced`/resolution mutants of earlier
   families) — they MUST fail.
4. **No semantic drift.** The S4 rules as accepted in round 3 (`[]` default, explicit null,
   allowlist warning, §2.3 surfacing before lock publication for install and
   update, `s4-warn`/`s4-enforce`) and the landed S6/E2/E4 rules are all
   still stated exactly once, with consistent closed-set spellings across
   prose, tables, schema, vectors and CLI.

## Verdict
Record `TASK-260917-1sz0b4_review-verdict-rev1.md` (an outcome resource on THIS
task, TASK-260917-1sz0b4) with the per-file merge table, transcripts and the verdict
`accept-landing` or `changes_requested` (concrete corrections). This task
carries no Change Request: record the verdict resource, tick the checklist
items you verified, then set this task's status yourself —
`task-board m 'set_status(TASK-260917-1sz0b4, status=done)'` on accept-landing, or
`status=to-dev` with the corrections listed on changes_requested. Do not touch
`TASK-260910-2ohnjo` (its revision 4 is already accepted; the orchestrator
lands PR #63 and closes it with `integrate_external`).
