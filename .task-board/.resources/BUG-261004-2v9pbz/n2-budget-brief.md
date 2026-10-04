# THE ONLY CURRENT INSTRUCTION — BUG-261004-2v9pbz (N2): repeated blobs bypass the expanded-snapshot budget (developer, code)

Source: docs/security-audit-2026-10-inline.md §N2 (issue #106). Acceptance criteria are on the element; satisfy every one.
1. RED FIRST through AdmitLocal with a real Git repo and a small MaxExpandedBytes: one blob at 64 paths over budget must be refused (admitted today); the one-blob and 64-distinct-blob controls keep their outcomes; add a repeated-tree (tree DAG) fan-out case.
2. Fix: aggregate accounting of emitted files, expanded bytes and canonical framing size, checked BEFORE each copy/append; keep unique-object limits; bound visited expanded tree entries; check context cancellation during the walk. The network admission path (proveRepository) must enforce the same bound: show it with a shared-path test or a network-entry regression.
3. No change to canonical snapshot bytes for in-budget repositories (run the targeted buildrepo conformance tests).
4. Mutants: move the check after the append; account only unique objects. Both must be killed.

## Produce mode (operator 2026-10-04, binding)
Maximise useful output now. Write the code and targeted tests for the packages you touch, and run ONLY those targeted tests (GOFLAGS=-work). Do NOT run the full local suite: the host has exec stalls, and the hosted gate on the Change Request is the arbiter. State plainly which claims are verified locally and which are left to the hosted gate. Never edit LOGBOOK.md or CHANGELOG.md. If a command hangs for more than 5 minutes, wait instead of retrying in a loop.
Then `task-board handoff BUG-261004-2v9pbz --role developer` and END YOUR TURN.
