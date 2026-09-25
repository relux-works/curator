# Review note — TASK-260924-10d3l1 revision 4 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Revision 3 was CHANGES_REQUESTED (F1: stale candidate reverting 60 trunk files). Revision 4 is the repair. Verify: (1) candidate tree vs
its recorded base changes EXACTLY internal/install/draftevidence_test.go (git diff --name-only base tree, excluding .task-board);
(2) that file's change equals the accepted content (per-file `git patch-id --stable` c7ce917d… of rev1/rev2, or — where trunk changed the
file (m28s6b) — both sides present and nothing of trunk dropped); (3) no CHANGELOG, no stray files; (4) validation log green. Focused
checks only; `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1` if needed. accept_cr or changes requested. No LOGBOOK.md.
