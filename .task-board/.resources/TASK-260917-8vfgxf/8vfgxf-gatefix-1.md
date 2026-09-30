# TASK-260917-8vfgxf — gate fix 1 (THE ONLY CURRENT INSTRUCTION, with 8vfgxf-brief.md)

Rev1 (tree decfbea0) failed the gate (run 36651575754) for two diagnosed reasons. Fix exactly these:
1. Windows checkout fails with `error: invalid path 'internal/opaquescan/nul.go'`. NUL is a reserved device name on Windows, so no file
   or directory may be named nul, con, prn, aux, com1–9 or lpt1–9, with any extension. Rename the file, e.g. to `nulopaque.go`, and
   check that no other new path uses a reserved name.
2. internal/envprofile TestManagerOwnedAbsenceReadsAreGuarded (state_read_guard_test.go:215):
   `internal/audit/audit.go:detectWithOpaquePaths tests not-exist after os.Stat without a seam route or reviewed allowlist reason`, and
   `reviewed allowlist entry has no matching absence-sensitive reader: internal/audit/audit.go:detect`.

   Your refactor renamed or moved the reader. Route the absence-sensitive read through internal/stateread (preferred). If it is not a
   manager-owned state read, move the reviewed allowlist entry to the new function name and keep its reason accurate. Do not widen the
   allowlist.

Run `go test ./internal/envprofile -run TestManagerOwnedAbsenceReadsAreGuarded` and `go test ./internal/audit ./internal/opaquescan`,
record the real exit codes, and run `git ls-files | grep -iE '(^|/)(nul|con|prn|aux|com[1-9]|lpt[1-9])(\.|/|$)'`, which must print
nothing. Set status development, update the results, run `task-board handoff TASK-260917-8vfgxf --role developer`, then END YOUR TURN.
No CHANGELOG/LOGBOOK edit.
