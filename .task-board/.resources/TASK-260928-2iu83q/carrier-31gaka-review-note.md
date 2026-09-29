# Review note — TASK-260928-2iu83q carrier of TASK-260927-31gaka (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

31gaka rev2 (tree 31e7a36f on base 86552087) was ACCEPTED (your rev2 re-review: shipped Codex seed revision A, B kept behind the switch).
It cannot land through its own Story (checkpointed on a discarded branch; #319 family). This carrier re-applies exactly that content on
trunk d8e87bac: rev2 tree 40f30ed4, gate green, 7 paths. Orchestrator checks: rev1 (3e740af5) matched the accepted content line-for-line;
rev2 fixed .github/ci/conformance-gaps.tsv so that its diff vs trunk is ONLY the nine TASK-260927-1e5qqm revision-B rows.
Verify: (1) for the 6 code/test paths, `git diff d8e87bac 40f30ed4` equals the accepted 31gaka delta modulo trunk context (no behaviour
beyond revision A; B behind the switch; seed record/marker rules; managed writes through E5's nofollow helpers); (2) the ledger diff is
exactly the 9 B rows; (3) `go test ./internal/envprofile -run 'Seed|Codex|Mcp|Status|Nofollow|Guarded'` and `go test ./internal/config`
pass with real exit codes; mutants from the 31gaka review (warning removed; strip under A) still killed. accept_cr or changes requested with
file:line. No LOGBOOK.md.
