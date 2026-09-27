# TASK-260910-2n0233 — gate fix (THE ONLY CURRENT INSTRUCTION, with 2n0233-sec-brief.md)

Revisions 1–2 fail on every lane on ONE thing: internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded (state_read_guard_test.go:221):
"internal/registry/boundary.go:ReadOnlyStateDir tests not-exist after os.Stat without a seam route or reviewed allowlist reason". The
deny-by-default §8.4.1 guard (Story 1a2i5a) scans your new code. Route that read through the internal/stateread seam (absent vs unreadable;
an unreadable high-water state must NOT be treated as absent — for a security boundary check it must fail closed), and add a row for the
unreadable case. Do not allow-list it unless it genuinely cannot be unreadable (then give the reason).
Also: trunk moved to `55b94af2` — `git fetch origin main`, combine (keep both sides), refresh-candidate if publication would refuse; VERIFY no
trunk revert. `task-board m 'set_status(TASK-260910-2n0233, status=development)'` first; bounded runs (the guard test + registry package, split);
append "Revision 3 — seam route for the high-water read", `resource update`, handoff; stay in the turn. If the loop detector refuses, stop
and report. No CHANGELOG/LOGBOOK edit.
