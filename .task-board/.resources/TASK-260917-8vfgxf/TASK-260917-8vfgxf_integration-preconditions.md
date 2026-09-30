# Integration preconditions — TASK-260917-8vfgxf rev3 (accepted)

Bound run: developer (implementer), CR-TASK-260917-8vfgxf-3 revision 3. Board left untouched at integrating; no handoff, no integrate invocation — landing is the runners step.

## Preconditions confirmed
- Branch: task-board/story/STORY-260928-9vt338; HEAD 0e3169bb (no commits past checkpoint; rev3 content fully uncommitted in working tree).
- Task TASK-260917-8vfgxf status: integrating. Story STORY-260928-9vt338 status: integrating.
- Working tree = 10 paths, matching accepted rev3: 6 modified (internal/audit/audit.go, internal/audit/audit_test.go, internal/contextaudit/contextaudit.go, internal/contextaudit/contextaudit_test.go, internal/envprofile/envprofile.go, internal/envprofile/state_read_guard_test.go) + 4 untracked (cmd/curator/nul_opaque_test.go, internal/envprofile/nul_opaque_test.go, internal/install/nul_opaque_test.go, internal/opaquescan/nulopaque.go).
- No Windows-reserved names: git ls-files grep for (nul|con|prn|aux|com[1-9]|lpt[1-9]) prints nothing.
- No CHANGELOG/LOGBOOK edits in working tree.
- No board writes made by this run (read-only queries only); no stray files beyond the 10 rev3 paths.

## Fresh verification (this run, real exit codes)
- go test -count=1 ./internal/opaquescan ./internal/audit ./internal/contextaudit → exit 0 (audit ok, contextaudit ok, opaquescan has no test files).
- go test -count=1 -run NUL|Opaque ./internal/install ./internal/envprofile ./cmd/curator → exit 0 (all three ok).
- go test -count=1 -run TestManagerOwnedAbsenceReadsAreGuarded ./internal/envprofile → exit 0.
- go vet ./internal/opaquescan ./internal/audit ./internal/contextaudit → exit 0.

## For the runner
Working tree is ready to land as-is. Nothing was committed, stashed, or reverted by this run.
