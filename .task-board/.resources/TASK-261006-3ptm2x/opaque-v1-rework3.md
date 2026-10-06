# THE ONLY CURRENT INSTRUCTION — TASK-261006-3ptm2x rework 3 (developer, code). One finding.
The reviewer (sol high, RUN-261006-9289ae) confirmed that the rev3 findings F1, F2 and F3 are fixed. ONE P1 bypass remains; read `TASK-261006-3ptm2x_review-verdict-rev4.md` finding 1 IN FULL, with its probe:
- In `internal/envprofile/store_boundary.go:211`, the commit-pinned v1 context re-check returns from the `member.Commit` branch BEFORE the full-snapshot opaque guard runs, so it computes and trusts v1 identities over NUL.
Fix: every v1 context-identity compute or trust path, commit-pinned included, runs the full-snapshot NUL guard first and refuses. Sweep the sibling branches in that file and in `internal/contextlock`/`contextaudit` for the same early-return pattern, and list each branch you checked.
Turn the reviewer's probe into a production-entry regression test, and name the mutant (moving the guard below the Commit return) that it kills.
Host rules: targeted tests only through `mini-build-lock` (envprofile has no fake executables; cmd/curator and install stay hosted). No LOGBOOK or remote-gate.sh edits.
Then `task-board handoff TASK-261006-3ptm2x --role developer` and END YOUR TURN.
