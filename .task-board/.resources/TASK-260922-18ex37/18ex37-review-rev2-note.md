# Review note — TASK-260922-18ex37 SPEC_PIN move + gap ledger fill, CR revision 2 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Story final revision of STORY-260922-2goxjs (carries the checkpointed gap ledger TASK-260922-3bbvrs). Review against `18ex37-brief.md`
and the task description; disposable clone.
0. Revision 1 was CHANGES_REQUESTED only for the stray root `TASK-260922-3bbvrs_results.md`; confirm revision 2 differs only by its removal (patch-id per file), then do the FULL review of items 1-4 (not done last cycle).
1. Pin: every SPEC_PIN / conformance-root pin location moved together to curator-spec `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (spec
   main 3d2c611 + erratum PR #88), exactly as the rc.12 promotion (48da2690) moved them.
2. Classification: `scriptHostExecutionPolicySections` (and any other bidirectional section/family classification) covers the new
   sections — `executable_identity_cases` + its definition — consumed or recorded unreachable with reason; nothing weakened.
3. Every new-root failure is fixed-by-classification or a gap-ledger row attributed to the owning board element with the spec commit that
   published the case verified; tally before/after in results; no driven row regressed to skip; ratchet intact.
4. Hosted gate green on the candidate (validation log).
accept_cr or changes requested with file:line. No LOGBOOK.md.
