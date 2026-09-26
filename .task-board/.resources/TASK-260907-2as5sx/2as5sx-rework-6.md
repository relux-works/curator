# TASK-260907-2as5sx — rework 6: Windows-only failure of your new test (THE ONLY CURRENT INSTRUCTION)

Revision 8 (refresh onto a48f584c) gate run 35994653101:
- windows-latest: `internal/install TestLockedNetworkRepositoryDistinguishesAbsentAndUnreadableCheckouts/unreadable_checkout_stops_fallback`
  draftsources_test.go:65: `lockedNetworkRepository error = <nil>, want typed manager_state_unreadable error`. Your fixture makes the
  checkout unreadable with a mechanism Windows does not honour (POSIX mode bits). Make the unreadable state real on Windows too (e.g. a
  non-directory where the checkout directory is expected, or a deny ACL) so the row asserts the typed error on all three OSes. If the
  production code cannot see "unreadable" on Windows at all, that is a production defect: fix it and say so. No skip, no ledger row.
- Race(ubuntu) `FAIL internal/install 1800.100s` is the known budget flake BUG-260922-3v8k23 (being fixed separately) — NOT yours; do not
  touch timeouts.
1. `task-board m 'set_status(TASK-260907-2as5sx, status=development)'` if needed. 2. Fix, bounded local run of the package's changed rows
   (+ `GOOS=windows go vet ./internal/install`). 3. No CHANGELOG edit. Artifacts only in $TMPDIR. 4. Append "Revision 9 — Windows unreadable
   checkout", `resource update`, `task-board handoff TASK-260907-2as5sx --role developer`. A write-boundary `policy warn` block is a warning.
