# TASK-260906-1xbrz6 — review verdict, revision 2 (CR-TASK-260906-1xbrz6-2)

**Verdict: ACCEPTED**

## What I checked
- The worktree's tree equals the candidate tree 677eb95b (checked with a temporary-index write-tree). Base is 05053cd.
- Diff read in full for env §9.4/§9.5/§9.6, manager §12.3 and the CHANGELOG. All four sentences use the same diagnostic (`environment_surface_unmanaged_conflict`, fail closed per §8.3) and the same recovery ops (`profile sync/use --takeover`). Import recovery retries the activation, not the import, so the `profile_import_name_taken` problem is answered. `cli/curator.md` is unchanged and has no `--takeover` on the import or global rows. The editorial follow-up (TASK-260906-3o75d6) was not started here.
- The rev1 findings are fixed. F1: all four sentences are pinned as exact whitespace-normalised sentences in their own sections (`validate.py` `require_exact_takeover_sentence`). F2: complete H3 heading lines are matched, and a section ends at the next same-or-higher heading (`markdown_h3_section`).

## Mutants I ran myself (candidate copy, jsonschema venv, calling `validate_takeover_closed_set_text()`)
Baseline PASS. Each mutant below FAILS:
1. §9.5 "fail closed" → "fail open" — FAIL (9.5 pinned sentence)
2. §9.5 "outside" → "inside" — FAIL
3. §9.4 subject `global install` → `global remove` — FAIL (9.4)
4. Heading `### 9.5 Onboarding` → `### 9.5 Onboarding-renamed` — FAIL (no section)
5. Sentence split at "fail closed. on" — FAIL
6. (my own) exclusion sentence moved from §9.5 into §9.7 — FAIL
7. Carrier list widened with `profile import` — FAIL (carrier != pinned list)
8. `[--takeover]` added to the `cli/curator.md` import row — FAIL (cli row)
9. Manager §12.3 "retry activation rather than `profile import`" → "retry `profile import`" — FAIL (manager 12.3)
All files were restored afterwards.

- `python -B -m unittest test_validate.TakeoverClosedSetTextTests`: 21 tests OK, rc=0 (zsh, pipefail).
- Full `tools/validate.py` in the worktree (tree == candidate): "validated 62 schemas and 1124 vector files", rc=0. Note: copies made with `git archive` fail on byte fixtures because of export-subst, on the base as well, so that is an artifact of the copy, not a regression. I did not rerun the full unittest discovery (~20 min) or regenerate-check; I rely on the handoff gate run for those.
- No schema or vector change is needed. The only takeover references in conformance are write-kind cases in environments-write-nofollow; no family lists the carrying operations.

## Recovery reasoning
- Import activation: activation follows §9.1 from the installed lock, and `profile use/sync --takeover` rewrites exactly the lock-represented paths. So a takeover followed by a retry of the activation is sound; the text correctly does not promise that repeating the import works.
- Global ops: the question of when the lock is published relative to materialization is real, and §9.4 does not settle it (the §9.2 update ordering does not extend to it). The packet is raised correctly: the question is exact, alternatives 1, 2 and 3 cover the space (3 is ruled out by the decision), and recommendation 1 is sound. It is not a defect of this leaf.

## Non-blocking observation for the orchestrator
The §9.4/§9.5/manager sentences still state the global recovery ("run sync/use --takeover, then retry that operation") without a condition. That is only true under packet alternative 1. If the orchestrator picks alternative 2, those three sentences and the `TAKEOVER_EXCLUSION_*` pins must change together.
