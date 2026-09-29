# TASK-260910-32gki6 — gate fix (THE ONLY CURRENT INSTRUCTION, with 32gki6-sec-brief.md)

Rev1 (tree e061d684, base 213a53e5) fails on macOS/Windows/race (run 36445692083), two tests:
- internal/envprofile TestReservedSkillName (status_test.go:509): want environment_reserved_command_name, got `environment_repair_failed:
  revalidate pinned snapshot: extract pinned snapshot dddd…: not a git repository: …/profile-repos/github.com_example_evil`.
- internal/envprofile TestSurfacingUnreadableManifest (surfacing_test.go:366): the test's own `remove …/contexts/mcp/figma-devmode/<pin>/
  agent-mcp.json` fails "no such file" — the store entry is no longer where the test (and trunk) put it.
Diagnosis: your pin-hash step re-extracts the pinned snapshot from the SOURCE repository. environments.md §4 ("Store integrity is verified
against the pin, not the marker… for a git member the git tree object identity of the pinned commit (or the recorded snapshot tree hash the
lock carries)…; This check needs no home marker and runs at every env resolve") — recompute the STORE ENTRY's tree hash from its bytes and
compare it with the pin the lock records; env resolve must not need the source repo or network. Rebuild-from-revalidated-snapshot belongs
to the entry-class (b) repair path of a REAL mutating operation, not to the verification step. Also keep store paths exactly as trunk lays
them out (the second failure suggests you moved/rebuilt the entry during a read path). Keep the verification order (enclosing boundary →
entries → pin hashes → home currency) and every existing diagnostic reachable (reserved-name, unreadable manifest …).
Do not weaken any test. Run `go test ./internal/envprofile -run 'Reserved|Surfacing|Store|Boundary|Pin|Resolve|Guarded'` (split) with real exit
codes. `task-board m 'set_status(TASK-260910-32gki6, status=development)'` first; update results; handoff; END YOUR TURN. No CHANGELOG/LOGBOOK.
