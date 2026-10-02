# THE ONLY CURRENT INSTRUCTION — TASK-261001-3bsyvh re-apply onto current trunk → revision 4 (orchestrator brief, binding)

The Story workspace was discarded on purpose. Rev3 (accepted, on e87d488b) went stale: trunk 67d83539 (Muse 2yvag1 and others) changed `.github/ci/conformance-case-counts.tsv` and conformance-gaps/root-artifacts again, and `worktree converge` refused with a 3-way conflict.

Nothing is lost. These refs are in the control repo:
- `refs/campaign/rjxrgs-rev1-20261001` (fe2b5f61): the ORIGINAL accepted rjxrgs rev1, the normative content.
- `refs/campaign/3bsyvh-rev2-20261001` (4a3d18c7): your rev2 delta on bd126a9a, for reference.

Do:
1. In your fresh workspace (current trunk 67d83539), re-apply the rjxrgs content unchanged, for example `git diff bd126a9a 4a3d18c7 -- . ':!.task-board' | git apply --3way`.
   - For the three `.github/ci/*.tsv` files: take trunk's rows plus rjxrgs's added rows.
   - Recompute exact counts keyed by manifest digest with the established tooling/tests. Never hand-guess.
2. Prove identity:
   - Every non-tsv path is byte-identical to the 4a3d18c7 version, or differs only by trunk context (show which).
   - The tsv rows are the union, with recomputed counts.
3. Re-run the rjxrgs rows from its results, plus `go test ./internal/conformancecoverage -count=1`, with real exit codes.
4. Add a "Revision 4 (re-apply)" section to the results. Then run `task-board handoff TASK-261001-3bsyvh --role developer` and END YOUR TURN.

No content change. Never spell any employer name.

## Update for revision 4
Use the rev3 snapshot as the source: `refs/campaign/3bsyvh-rev3-20261001` (d540fbb1), the accepted rev3 delta on base e87d488b:
`git diff e87d488b d540fbb1 -- . ':!.task-board' | git apply --3way`

Trunk now also carries the Muse conformance rows: keep them, and recompute the counts exactly.
