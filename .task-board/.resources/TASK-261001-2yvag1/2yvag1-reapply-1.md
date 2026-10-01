# THE ONLY CURRENT INSTRUCTION — TASK-261001-2yvag1 re-apply onto current trunk → revision 6 (orchestrator brief, binding)

The Story workspace was discarded on purpose. Rev5 (green, on bd126a9a) went stale before review: trunk e87d488b (2tx81l, content-hash v2) changed conformance-case-counts.tsv and internal/envprofile/managed.go. `worktree converge` refused with a conflict on internal/conformancecoverage/content_hash_v2_gaps_test.go.

Nothing is lost. Your rev5 delta is `refs/campaign/2yvag1-rev5-20261001` (80fc6075) in the control repo, on base bd126a9a.

Do:
1. In your fresh workspace (trunk e87d488b), re-apply it: `git diff bd126a9a 80fc6075 -- . ':!.task-board' | git apply --3way`.
   Resolve conflicts keeping BOTH sides:
   - 2tx81l's content-hash v2 rows and gaps test;
   - your Muse rows.
   Recompute `.github/ci/conformance-case-counts.tsv` exactly (keyed by manifest digest) with the established tooling. Never hand-guess.
2. Prove that every non-conflicting path is byte-identical to 80fc6075, or differs only by trunk context. Explain each resolved hunk.
3. Re-run, with real exit codes:
   - `go test ./cmd/curator ./internal/envprofile ./internal/conformancecoverage -count=1`;
   - the Muse rows (16 link-state, 36 fragment-v3);
   - the six env-status/guard tests;
   - the 3b3oyi TestPosture* and subcommand --help rows.
4. Add a "Revision 6 (re-apply)" section to the results. Then run `task-board handoff TASK-261001-2yvag1 --role developer` and END YOUR TURN.

No product change beyond conflict resolution. Never spell any employer name. Do not touch askpass code (BUG-261001-2772iz).
