# Rework brief — TASK-260910-2ohnjo, revision 2 (answers review verdict rev1)

Read `TASK-260910-2ohnjo_review-verdict-rev1.md` (task outcome) first; it is the
authority for this round. Everything it marks PASS stays as it is. Work in the
same curator-spec Story worktree
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260910-1lf0m5/worktree`,
your rev1 edits are still there, uncommitted); rules in
`remediation-spec-producer-rules.md` (attach the spec diff as
`TASK-260910-2ohnjo_spec-patch_rev2.patch`).

## 1 — generator drift (high)
`tools/generate-vectors/manager_config.go:22` still emits
`"passable_env_names": nil`; the schema (`:454`) and §12.1 say `[]`.
`make regenerate-check` exits 2 and would revert the manager-config vector
bytes, drop the four new schema cases and move the manifest/release pins.
Change the generator to the new default, add the four schema cases to the
generator so they are produced (not hand-written), run `make regenerate`, and
make `make regenerate-check` exit 0 in the worktree (quote it).

## 2 — one consistent surfacing order (high)
§2.3 (`protocol/environments.md` ≈481–484) requires surfacing "before the lock
is published or any surface is (re-)materialized"; §9.2 update (≈1714–1717)
says "After publishing the new lock". Specify one sequence for install AND
update — recommendation: surfacing rows are printed after resolution and
before the lock is published and before any materialization, for both — and
cover the update order in a vector.

## 3 — S4 vectors must be semantically checked (high)
`tools/validate.py` only parses and integrity-hashes
`vectors/environments-env-passthrough.json`; the reviewer flipped the
enforcing absent-knob case to pass `FIGMA_API_KEY` and replaced expected
surfacing bytes with `INVALID\n` (pins refreshed) and `validate.py` still
exited 0. Add a dedicated consumer wired into the validation entry point
(next to `validate_environment_vectors`): recompute the effective
`passable_env_names` per profile/knob state, the passed/dropped/warned name
sets and diagnostics, and the surfacing output bytes from the declared inputs,
and compare with the expected fields; add unit tests in `tools/test_validate.py`
including a mutant like the reviewer's that MUST fail.

## 4 — contradictory negative fixture (medium)
`non-empty-allowlist-silent` is `conforming: false` with `diagnostic: null`
describing correct silence while the reason describes emitting the warning.
Make the observation match the verdict (or make silence positive and add an
emitting negative) and assert it through the consumer of item 3.

## Validation and handoff
`make validate` and `make regenerate-check` (repo venv on PATH,
`set -o pipefail`, quote outputs and exit codes). Attach
`TASK-260910-2ohnjo_spec-patch_rev2.patch`, update
`TASK-260910-2ohnjo_evidence.md` (closure of 1–4 with file:line, transcripts),
tick the checklist items you satisfy, then
`task-board handoff TASK-260910-2ohnjo --role doc-writer`.
