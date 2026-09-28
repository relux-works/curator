# TASK-260927-4pv4au — rework 1 (THE ONLY CURRENT INSTRUCTION, with 4pv4au-brief.md)

Review rev2 = CHANGES REQUESTED (`TASK-260927-4pv4au_review-verdict-rev2.md`). Accepted: one switch, permissive default, loadConfig warns
once on stderr, `status --json` security_posture_rows (spec-required, manager.md §10). Blocking F1: cmd/curator/main.go:214-216
(runEnforcedShim) still prints `warning: security_posture_permissive…` into the launched script's stream, and rev2 hid the failure by
switching the fixtures of internal/scriptworker/derive_test.go:1138 and internal/install/scriptpolicy_test.go:671 to `security_posture:
hardened`. Do exactly:
1. Remove the emission from runEnforcedShim (the launched command's stdout/stderr belong to the caller's pipeline; manager.md §7.1 names
   status reporting, not the launcher shim).
2. Restore the ORIGINAL schema-1/permissive fixtures in both tests (a hardened variant may be ADDED, not substituted).
3. Add a negative test: a permissive (and a schema-1) launch through the enforced shim produces no `security_posture_permissive` bytes on
   the launched command's stdout or stderr; show the mutant (emission re-added) fails it with a real exit code.
Rebase: trunk is now 86552087 (ryh3kw) — `git fetch origin main` only if a conflict blocks you; the orchestrator will carry at landing.
Set status development; update the results resource ("Revision 3 — no warning on the launch path"); handoff; END YOUR TURN.
No CHANGELOG/LOGBOOK edit.
