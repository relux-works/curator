# Review note — TASK-260922-cww1ov revision 10 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

0017 credential modes (Story STORY-260922-1cenbr, story_final with F-C1 1t551d / F-C2 1t2w1q checkpoints). Revision 5 was ACCEPTED on
content (see earlier verdicts); revision 10 = refresh onto trunk faf509ae (`cww1ov-refresh-9.md`) + a successor's gate fix. Verify:
0. By rule producers never edit LOGBOOK.md — the candidate changes LOGBOOK.md vs its base: if these are entries this Story added, that is
   a BLOCKING finding (revert to trunk bytes); name the hunk.
1. Diff against the last accepted content (rev5 patch) and name every difference; each must be (a) trunk content combined correctly
   (both sides present), (b) a migration of a trunk-added state reader onto the stateread seam (R5 script-worker readers, lock replay in
   internal/install/draftsources.go, runtimestore) — check each is correct and covered, or (c) the successor's gate fix — judge it.
2. No revert of trunk: git diff --name-only base..tree == this Story's paths only; CHANGELOG.md equals trunk; no stray files.
3. 0017 rows (hazards, migration/recovery, no-copy scan) still pass; kill one mutant of your choice in a newly migrated reader (bounded,
   focused tests only — host memory is tight). Hosted gate green on all lanes.
accept_cr or changes requested with file:line. No LOGBOOK.md.
