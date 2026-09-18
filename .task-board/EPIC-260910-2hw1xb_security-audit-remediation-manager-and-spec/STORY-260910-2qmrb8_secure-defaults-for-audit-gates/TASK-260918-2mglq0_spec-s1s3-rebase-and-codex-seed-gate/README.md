# TASK-260918-2mglq0: spec-s1s3-rebase-and-codex-seed-gate

## Description
Revision 4 of the accepted S1+S3 specification (TASK-260910-2qtiho, revision 3 accepted on curator-spec base e8b53a0): the orchestrator rebased the Story worktree onto curator-spec main 1ca4b3d (E3, E6 landed) as a mechanical union; this task absorbs the E3 codex-seed shipped-revision row into the closed posture inventory (thirteen rows) in manager section 10, environments section 12, the security-posture vectors and validator, and CHANGELOG, keeping everything else byte-identical to the rebased revision 3. Reviewed as a spec revision; landed as one curator-spec PR that also closes TASK-260910-2qtiho.

## Scope
(define task scope)

## Acceptance Criteria
1. Story worktree state equals the attached rebased revision-3 patch before the edits (patch-id 43f07dee) and the new patch is git diff HEAD on base 1ca4b3d. 2. manager section 10 posture table lists codex-seed (A or B, section 7.4, shipped) at a fixed stated position; every twelve becomes thirteen; environments section 12 lists it and states the per-home codex_seed_record rows are not posture rows. 3. security-posture.json outputs pin the row; validator expected rows, tests and a rule-7 negative cover it. 4. E6 needs no new row (store-boundary reference checked). 5. make regenerate, regenerate-check, validate exit 0; EMPTY curator delta; evidence Revision 4 section.
