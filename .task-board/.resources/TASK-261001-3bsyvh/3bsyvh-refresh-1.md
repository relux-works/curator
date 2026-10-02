# THE ONLY CURRENT INSTRUCTION — TASK-261001-3bsyvh base refresh → revision 2 (orchestrator brief, binding)

Revision 1 (identity carrier of accepted rjxrgs rev1) was ACCEPTED, but landing refused with integration_base_moved: trunk (now bd126a9a) changed `.github/ci/conformance-case-counts.tsv`, which rev1 also changes. The acceptance is released. The content is not in question; this is a base refresh only.

1. Refresh onto the current trunk: `task-board worktree refresh-candidate TASK-261001-3bsyvh`.
   - On a conflict, follow `--replay-resolutions`. Never hand-commit the replay worktree.
   - The counts file must carry both sides' rows with EXACT counts, keyed by manifest digest. Recompute them from the suites (the established count tooling or tests); never hand-guess.
2. Prove the refreshed candidate is identical to rev1 apart from the merged counts file and any trunk-only paths. Report `git diff` stats per file and the resolution in the results.
3. Re-run with real exit codes:
   - the rjxrgs rows named in its results (mixed receipt marker and lifecycle);
   - `go test ./internal/conformancecoverage -count=1`;
   - any counts/ledger consistency check the gate uses.
4. Add a "Revision 2 (refresh)" section to the results. Then run `task-board handoff TASK-261001-3bsyvh --role developer` and END YOUR TURN.

No product change beyond the conflict resolution. Never spell any employer name.
