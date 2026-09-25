# Review note — TASK-260922-18ex37 SPEC_PIN move + gap ledger fill, CR revision 1 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Story final revision of STORY-260922-2goxjs (carries the checkpointed gap ledger TASK-260922-3bbvrs). Review against `18ex37-brief.md`
and the task description; disposable clone.
0. BLOCKING by rule: the candidate contains the root file `TASK-260922-3bbvrs_results.md` (came in with the 3bbvrs checkpoint) —
   task documents are board resources, never repository files; record it (removal required) and continue the full review. Also check
   for any other stray artifacts (`test/`, `ledger/`, root `TASK-*`/`BUG-*` files).
1. Pin: every SPEC_PIN / conformance-root pin location moved together to curator-spec `dcc7f015e2d97edf2d52928afb6fd79ec8129e8b` (spec
   main 3d2c611 + erratum PR #88), exactly as the rc.12 promotion (48da2690) moved them.
2. Classification: `scriptHostExecutionPolicySections` (and any other bidirectional section/family classification) covers the new
   sections — `executable_identity_cases` + its definition — consumed or recorded unreachable with reason; nothing weakened.
3. Every new-root failure is fixed-by-classification or a gap-ledger row attributed to the owning board element with the spec commit that
   published the case verified; tally before/after in results; no driven row regressed to skip; ratchet intact.
4. Hosted gate green on the candidate (validation log).
accept_cr or changes requested with file:line. No LOGBOOK.md.
