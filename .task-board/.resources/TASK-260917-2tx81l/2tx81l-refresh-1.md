# THE ONLY CURRENT INSTRUCTION — TASK-260917-2tx81l base refresh → revision 4 (orchestrator brief, binding)

Revision 3 was ACCEPTED (astra), but landing refused with integration_base_moved: trunk (now bd126a9a) changed `.github/ci/conformance-case-counts.tsv`, which rev3 also changes. The acceptance is released. The content is not in question; this is a base refresh only.

1. Refresh onto the current trunk: `task-board worktree refresh-candidate TASK-260917-2tx81l`.
   - On a conflict, follow `--replay-resolutions`. Never hand-commit the replay worktree.
   - The counts file must carry both sides' rows with EXACT counts, keyed by manifest digest. Recompute them from the suites (the established count tooling or tests); never hand-guess.
2. Prove the refreshed candidate is identical to rev3 apart from the merged counts file and any trunk-only paths. Report `git diff` stats per file and the resolution in the results.
3. Re-run with real exit codes:
   - your content-hash v2 rows;
   - `go test ./internal/conformancecoverage -count=1`;
   - any counts/ledger consistency check the gate uses.
4. Add a "Revision 4 (refresh)" section to the results. Then run `task-board handoff TASK-260917-2tx81l --role developer` and END YOUR TURN.

No product change beyond the conflict resolution. Never spell any employer name.
