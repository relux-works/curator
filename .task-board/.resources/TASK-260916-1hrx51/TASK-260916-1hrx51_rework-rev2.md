# Rework brief — TASK-260916-1hrx51, revision 2 (answers review verdict rev1)

Read `TASK-260916-1hrx51_review-verdict-rev1.md` (task outcome) first; it is the
authority for this round. Everything it marks present/passing stays exactly as
it is (admission semantics, knobs, vectors, generator, schema cases). Work in
the same curator-spec Story worktree
(`/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-260916-2d9coh/worktree`,
your rev1 edits are still there, uncommitted); rules in
`remediation-spec-producer-rules.md` (attach the spec diff as
`TASK-260916-1hrx51_spec-patch_rev2.patch`).

## R1 — manager §1 closed lock set (required)
`profiles/manager.md` §1 (≈ lines 51–62) enumerates the keys `locked` "contains
only" and closes with "No other `environments` knob is lockable or carried by
the system file." Add `environments.transitive_system_modules` to that closed
enumeration (spelling identical to `protocol/environments.md` §12.2 and
`schemas/v1/system-config-v2.schema.json`), with the direction restriction
(`error` only) stated where the other directional keys state theirs
(`isolation` → `shared`). Nothing else in the set changes.

## R2 — RFC 2119 formulation (required)
Rewrite the admission block `protocol/environments.md` §3 (≈ 510–535) and the
direction rule in §12.2 (≈ 2352–2355) with explicit implementer obligations:
materialized output MUST contain only admitted system modules; under `drop` a
transitive system module MUST be skipped and the manager MUST emit
`context_system_module_dropped` naming package and module; under `error`
resolution MUST fail with `context_system_module_transitive` and MUST NOT
write or change the lock; a system file MUST lock `transitive_system_modules`
only to `error` and `system_module_waivers` MUST NOT be lockable. Keep the
existing sentences' meaning; only the modality changes.

## Evidence fix
The rev1 evidence says eleven schema cases; the index adds seven cases plus
four expected byte files — correct the count.

## Validation and handoff
`make validate` (repo venv on PATH, `set -o pipefail`, quote outputs and exit
codes) and `make regenerate-check` if it applies to your changes. Attach
`TASK-260916-1hrx51_spec-patch_rev2.patch`, update
`TASK-260916-1hrx51_evidence.md` (R1/R2 closure with file:line, transcript),
tick the checklist items you satisfy, then
`task-board handoff TASK-260916-1hrx51 --role doc-writer`.
