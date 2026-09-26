# Review note — TASK-260924-20o9dk revision 5 (carry-forward; orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 4 was ACCEPTED on content. Revision 5 = converge onto trunk 3bdcfe07 with a 3-way merge on the intersecting paths
internal/install/draftsources.go and internal/snapshot/capture.go / capture_test.go (cww1ov migrated readers there). Verify: every other path
per-file identical to rev4 (patch-id; renames compare by new path); on the three intersecting paths BOTH sides are present — cww1ov's
stateread migration AND 20o9dk's changes (C1 object-format check, #90 rules, 11burj's replay recovery) — nothing dropped or duplicated;
no revert of trunk; validation log green on all lanes. Focused checks only (host memory); run the draftsources replay rows and the
snapshot capture tests once. accept_cr or changes requested with file:line. No LOGBOOK.md.
