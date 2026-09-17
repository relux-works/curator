# Review brief — TASK-260917-1972im: landing review of the rebased E4 spec tree

Round 2 (`TASK-260916-1x0ogh_review-verdict-rev2.md`) ACCEPTED the E4 candidate
(`TASK-260916-1x0ogh_spec-patch_rev2.patch`, base curator-spec `07e2b41`).
Since then curator-spec `main` moved to `f544a01` (S6 #60 and E2 #61 landed)
on overlapping paths, so the orchestrator rebased the accepted candidate for
landing. The board Change Request stays accepted (empty curator delta); this
round reviews the EXACT tree that will land:

- delivery worktree (read-only for you):
  `/Users/administrator/Developer/ReluxWorks/.worktrees/curator-spec-e4-trust-roots`
  at commit `0da40207a70d0b6c990c8f1bc8c79e59f218bf6d` = head of
  https://github.com/relux-works/curator-spec/pull/62 (branch
  `spec/e4-provider-trust-roots`, parent `f544a01`).

## What to verify
1. **Merge fidelity.** `git diff f544a01..0da4020` vs the accepted patch: for
   every file, either the hunks are byte-identical / context-only, or the file
   is one of the merged sources — `protocol/environments.md` (§12 status
   paragraph, §12.2 lockable list), `profiles/manager.md` (§1 lock set),
   `schemas/v1/system-config-v2.schema.json` (lockable enum + properties),
   `tools/generate-vectors/{manager_config,system_config,manager_config_test}.go`,
   `tools/validate.py`, `tools/test_validate.py` — or a regenerated file
   (manifest, release pins, schema cases, manager-config vector). For each
   merged source confirm the result is the UNION of the landed S6/E2 content
   (compare with `f544a01`) and the accepted E4 content (compare with the
   rev2 patch): nothing landed was dropped, nothing of E4 was dropped, no
   third content was invented. Quote file:line.
2. **Regeneration exactness.** `make regenerate-check` in a disposable byte
   copy with a temporary git baseline (as in earlier rounds) exits 0; the
   regenerated schema cases carry all three knobs (`transitive_system_modules`,
   `system_module_waivers`, `provider_directories`) consistently.
3. **Validation.** `make validate` (repo venv on PATH, `set -o pipefail`) —
   quote the three gate outputs and exit code. Re-run your round-1 mutant
   (`s6-planted-path-provider-warns-then-refuses.revision_b` resolved instead
   of refused) — it MUST fail.
4. **No semantic drift.** The E4 rules as accepted in round 2 (revision A
   keeps PATH + warns; revision B trust roots only; disjoint outcomes;
   unreadable ≠ absence; posture rows) and the landed E2/S6 rules are all
   still stated exactly once, with consistent closed-set spellings across
   prose, tables, schema, vectors and CLI.

## Verdict
Record `TASK-260917-1972im_review-verdict-rev1.md` (an outcome resource on THIS
task, TASK-260917-1972im) with the per-file merge table, transcripts and the verdict
`accept-landing` or `changes_requested` (concrete corrections). This task
carries no Change Request: record the verdict resource, tick the checklist
items you verified, then set this task's status yourself —
`task-board m 'set_status(TASK-260917-1972im, status=done)'` on accept-landing, or
`status=to-dev` with the corrections listed on changes_requested. Do not touch
`TASK-260916-1x0ogh` (its revision 2 is already accepted; the orchestrator
lands PR #62 and closes it with `integrate_external`).
