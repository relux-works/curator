# Review note — TASK-260922-cww1ov revision 11 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

0017 credential modes (Story STORY-260922-1cenbr, story_final with F-C1 1t551d / F-C2 1t2w1q checkpoints). Revision 5 was ACCEPTED on
content (see earlier verdicts); revision 11 = LOGBOOK revert + refresh onto trunk 0a628621 (`cww1ov-rework-11.md`); rev10 = refresh onto faf509ae + a successor gate fix. Verify:
0. By rule producers never edit LOGBOOK.md — the candidate changes LOGBOOK.md vs its base: if these are entries this Story added, that is
   a BLOCKING finding (revert to trunk bytes); name the hunk.
1. Diff against the last accepted content (rev5 patch) and name every difference; each must be (a) trunk content combined correctly
   (both sides present), (b) a migration of a trunk-added state reader onto the stateread seam (R5 script-worker readers, lock replay in
   internal/install/draftsources.go, runtimestore) — check each is correct and covered, or (c) the successor's gate fix — judge it.
2. No revert of trunk: git diff --name-only base..tree == this Story's paths only; CHANGELOG.md equals trunk; no stray files.
3. 0017 rows (hazards, migration/recovery, no-copy scan) still pass; kill one mutant of your choice in a newly migrated reader (bounded,
   focused tests only — host memory is tight). Hosted gate green on all lanes.
accept_cr or changes requested with file:line. No LOGBOOK.md.

ADDENDUM for revision 11: LOGBOOK.md is now absent from the candidate (verify). New vs rev10: internal/conformancecoverage/coverage.go and
coverage_test.go — name why this Story changes them (a trunk-added reader migrated onto the stateread seam after 18ex37 landed?) and judge.
Points 1–3 of this note were NOT verified last round — verify them fully now.
