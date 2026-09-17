# Rework brief — TASK-260916-1qfpu4, revision 2 (E5 nofollow write discipline)

Revision 1 was rejected with two corrections
(`TASK-260916-1qfpu4_review-verdict-rev1.md`, F1–F2) plus one rule violation
the orchestrator adds (F3). Everything else passed; keep it byte-identical.

## Corrections (all required)
- **F1 — pin scenarios (high; producer rule 7).** `tools/validate.py`
  checks only the case-name inventory of `environments-write-nofollow.json`
  and then models every case generically, so an internally consistent
  passing case substituted under any required name survives (0/9
  rejected). Bind each required scenario to its discriminating inputs
  (operation, authorization/takeover flag, marker ownership, target kind and
  link destination, parent-link state, …) so the validator refuses a corpus
  in which a named case no longer exercises its branch; add tests in
  `tools/test_validate.py` that substitute internally consistent alternative
  cases under each retained name and require rejection through the validator
  entry point (`validate.main()`), not through inventory checks.
- **F2 — private destination links (medium).** §8.3.1 says the nofollow
  refusal also applies when a backup, marker or ledger destination traverses
  or names a link the manager did not create in this operation, but the
  validator routes every symlink target through the managed-surface
  foreign-manager logic. Model manager-private destinations (backup, marker,
  ledger) distinctly: a directly symlinked backup target with no parent
  symlink is `environment_write_would_follow_link` with the former target
  untouched — add that vector case, a negative test for the wrong
  `environment_foreign_manager_detected` diagnostic, keep the existing
  parent-traversal case, and state in the evidence what is vectorized vs
  text-only for marker/ledger destinations.
- **F3 — remove the curator LOGBOOK delta (rule).** The runtime Change
  Request for this task carries a `LOGBOOK.md` handoff entry in the CURATOR
  Story worktree (`<curator control root>/.temp/STORY-260916-73a5zg/worktree`).
  Campaign rules forbid LOGBOOK edits: run `git checkout -- LOGBOOK.md`
  there (the curator delta must be empty), keep findings in the evidence
  and task notes only, never edit any LOGBOOK.md again.

## Validation and handoff
`make validate` and the regeneration proof (exit codes); evidence "Revision
2" section; `TASK-260916-1qfpu4_spec-patch_rev2.patch` = `git diff HEAD` of
the curator-spec worktree (base `684c9f1`) with new files via `git add -N`;
`task-board handoff TASK-260916-1qfpu4 --role doc-writer`. Worktree and
rules unchanged.
