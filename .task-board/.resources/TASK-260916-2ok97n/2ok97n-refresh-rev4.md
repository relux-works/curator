# TASK-260916-2ok97n (R5) — publish revision 4 onto fresh trunk (THE ONLY CURRENT INSTRUCTION). HIGHEST PRIORITY.

Your revision-4 work (unresolved declared exec → absent + reported, no refusal; the System32 identity fix; rows + mutants)
is DONE and uncommitted in the Story worktree (safety ref `refs/campaign/r5-rev4-delta-20260924`, bdeaa1c8). Publication was
refused `change_request_base_authority_mismatch`: the Story checkpoint `ac0e9774` (refreshed onto fad88136) does not descend
from trunk `1511b345` (2kqa77 landed: 10 non-board paths, incl. `.github/ci/gate-selftest.sh`, `.github/workflows/ci.yml`,
CHANGELOG.md, tools/goreleaserconfig). The orchestrator has FROZEN all other curator landings until R5 lands.
1. `task-board m 'set_status(TASK-260916-2ok97n, status=development)'`.
2. Combine trunk's incoming content into the working candidate: `git diff fad88136 1511b345 -- . ':!.task-board' | git apply
   --3way` (never bring `.task-board/**`); keep both sides on overlaps; leave nothing staged.
3. `task-board worktree refresh-candidate TASK-260916-2ok97n`; replay conflicts only via its `--replay-resolutions` template.
4. Prove refreshed candidate = your rev-4 work + trunk's incoming, nothing else. Bounded re-runs:
   `go test ./internal/scriptworker/... ./internal/install -run 'Script|Enforced'`, `sh .github/ci/ledger-consistency.sh`,
   `sh .github/ci/gate-selftest.sh`.
5. Append "Revision 4 — refresh onto 1511b345" to results, `task-board resource update` the results file, then
   `task-board handoff TASK-260916-2ok97n --role developer`; stay in the turn while the gate runs. A
   `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review` and that revision 4 was published.
