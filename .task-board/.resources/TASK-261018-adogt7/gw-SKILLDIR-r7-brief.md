# THE ONLY CURRENT INSTRUCTION — TASK-261018-adogt7: rework round 7 of W-SKILLDIR (developer)

Review B on CR revision 2 requested changes. Read the attached verdict `TASK-261018-adogt7_review-verdict-rev2.md` and its probe `TASK-261018-adogt7_review-cache-probe-rev2.py`. The orchestrator has converged your Story onto fresh trunk (it now includes PR #145 CODEX162 and PR #146 MODES); check `task-board worktree status` first.

## Finding `audit-markdown-cache-reuse` (blocking, AC 4)

`internal/audit/audit.go` adds Markdown scanning, but `RulesetVersion` (line ~53) stays at 4, the same as the CR base. The cache-hit early return (line ~384) therefore reuses an allowance recorded before the change and skips the new detector. A package whose Markdown declares network use bypasses the selected-package network declaration gate. `docs/profile-skill-directories.md` wrongly claims that ruleset 4 prevents earlier allowances.

## Do this

1. Bump `RulesetVersion` so that every cache entry from before the Markdown detector misses the cache. Check whether anything else keys on that version (cache paths, tests, goldens, docs) and update it consistently.
2. Turn the reviewer's probe into a named regression test in the repository: an allowance seeded under the old ruleset must not satisfy the gate for a package whose Markdown now trips the detector. Pair it with the positive control: a fresh allowance under the new ruleset passes for a clean package. Add a narrowing mutant that restores the old version number; it must fail the regression for that reason.
3. Fix the docs claim in `docs/profile-skill-directories.md` to state the actual invalidation rule.
4. Fix any other P0/P1 in the verdict exactly as named; P2 only where the fix is local and safe. Stay inside the original brief `gw-SKILLDIR-brief.md` (CIP-0021 scope, no widening).
5. Run the affected tests, `go build ./...`, `go vet ./...`. rc.7 wave rule: `CHANGELOG.md` and the root `README.md` equal trunk. Do not run `refresh-candidate` unless `worktree status` reports a state only it resolves. Do not add checklist items about the post-turn validation.

## Hand-off

1. Add a "Rework round 7" section to `TASK-261018-adogt7_gaps.md`: each verdict item and how it was fixed, with file and line. Retitle any open "Stop-The-Line" or "Blocker" section from an earlier round as "Resolved in an earlier round".
2. Run `task-board handoff TASK-261018-adogt7 --role developer`; it must publish a new CR revision.
3. END YOUR TURN.
