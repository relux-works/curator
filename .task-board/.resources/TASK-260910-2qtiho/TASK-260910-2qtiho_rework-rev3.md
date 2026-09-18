# Rework brief — TASK-260910-2qtiho, revision 3 (S1 + S3)

Revision 2 closed F1 and the twelve-gate inventory; one correction remains
(`TASK-260910-2qtiho_review-verdict-rev2.md`): the provenance vocabulary.
The settled closed set is exactly `profile`, `explicit`, `lock`, `shipped`
(the rev-2 rework brief); the candidate spells `profile-default`, `explicit`,
`locked`, `shipped`. Rename mechanically and consistently: `profiles/manager.md`
§10 rows, `protocol/environments.md` §12 rows, CHANGELOG, every vector
value (`security-posture.json` provenance / profile_source / sources / row
values), the validator's `SECURITY_POSTURE_PROVENANCE` and derived expected
rows, and the tests; add negative checks rejecting the superseded spellings
(`profile-default`, `locked` as provenance values). Do NOT rename
configuration fields such as `system.locked` — those are not provenance
values. Nothing else changes.

## Validation and handoff
`make regenerate`, `make validate` and the regeneration proof (exit codes);
evidence "Revision 3" section; `TASK-260910-2qtiho_spec-patch_rev3.patch` =
`git diff HEAD` of the worktree (base `e8b53a0`) with new files via
`git add -N`; EMPTY curator delta; `task-board handoff TASK-260910-2qtiho --role doc-writer`.
