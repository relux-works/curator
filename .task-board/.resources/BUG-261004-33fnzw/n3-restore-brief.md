# THE ONLY CURRENT INSTRUCTION — BUG-261004-33fnzw (N3): unmanage restore widens a private file's mode (developer, code)

Source: docs/security-audit-2026-10-inline.md §N3 (issue #106). Acceptance criteria are on the element. N4 (BUG-261004-13ptlq, the symlink backup) follows in this same Story after you, so shape the restore plan to carry entry TYPE as well as mode.
1. RED FIRST, full CLI round-trip (profile install → profile use --takeover → env unmanage --restore-backups --env claude_code): 0600 stays 0600 (fails today: 0644); 0644 stays 0644; an executable mode is preserved where supported.
2. Fix: keep type and mode of each backup entry in the restore plan; protected atomic write that never widens the original permissions.
3. Unix permission assertions; state Windows ACL behaviour explicitly (tested or bounded with reason).
4. Mutant: restore with a hard-coded 0644 again; it must be killed.

## Produce mode (operator 2026-10-04, binding)
Maximise useful output now. Write the code and targeted tests for the packages you touch, and run ONLY those targeted tests (GOFLAGS=-work). Do NOT run the full local suite: the host has exec stalls, and the hosted gate on the Change Request is the arbiter. State plainly which claims are verified locally and which are left to the hosted gate. Never edit LOGBOOK.md or CHANGELOG.md. If a command hangs for more than 5 minutes, wait instead of retrying in a loop.
Then `task-board handoff BUG-261004-33fnzw --role developer` and END YOUR TURN.
