# Rework round 3 logbook

Deciding review finding: unrelated-takeover-argument-discard, row 12.
The common CLI boundary now refuses consent on every unrelated registered
command before help, wrappers or providers can discard it. The closed allowlist
and the existing operation parsers retain the five activating operations.

Only six CLI files changed relative to the round-2 signed snapshot. Native
rows 1–11, Muse/opencode/managed layouts and the root README are unchanged.
All 3,490 previous top-level test names are retained; four tests were added.
The current list-only narrowing plant is killed by both named list regressions.
All twelve retained native controls were freshly exercised and killed.
Historical survivors/failed plants are declared separately in the gaps file.

Exact signed gate candidate: 3e00fcc8abe75e411a88773c65c4042ae2448eb0, G.
Tree 2fe068da665b203ed48e52cfcb7fba74de96e37e equals the uncommitted worktree.
The Story HEAD remains checkpoint 5d8e6b07f7f485557d9bbb065c9692348ca065c2.
A fresh protected-trunk comparison has no path overlap.

Local validation is green: 85 packages, 28 covered CLI/envprofile groups,
3,494 unique top-level tests, 36 classified skips; lint/vet/build green.
Twenty-one groups were executed on this identity; seven internal groups and
83 unrelated packages reuse exact unchanged green evidence, recorded explicitly.
No real operator HOME was used as a write target.

Hosted run 39117060646 is green on attempt 2, watch exit 0, on the exact head.
Attempt 1 failed only two Complete runner cleanup steps after successful
repository steps; all eighteen other jobs passed. The accepted failed-job retry
passed both jobs including cleanup. Final provider status: 20 success and two
conditional skips. Windows artifact confirms all new and retained controls.
Fresh final integrity remains green. Outcome logs, hashes and coverage are
attached before the review handoff; no producer commit was made on Story.


## Second recovery: runtime signing remains blocked

No source edits. Four row-12 regression tests and build rerun directly, exit 0;
list-only narrowing mutant child exit 1 for both named list regressions.
Exact worktree identity still equals signed candidate 3e00fcc8abe75e411a88773c65c4042ae2448eb0;
fresh full-hosted provider read reports success on attempt 2.
CR revisions 2 and 3 both failed signing before CI, exit 1. The prior handoff
from a bootstrapped shell did not repair post-turn validation. Fresh shell-local
signing probe passes, exit 0. Do not keep republishing the same failing recovery.
Runtime owner must correct the validator signing environment; this run may not
restart shared services or broaden product scope to work around it.
Evidence: BUG-261018-uoisfe_recovery2-r3.md, regression-recovery2-r3.log,
cr-validation-failures-r3.txt. No control-root files directly edited.
