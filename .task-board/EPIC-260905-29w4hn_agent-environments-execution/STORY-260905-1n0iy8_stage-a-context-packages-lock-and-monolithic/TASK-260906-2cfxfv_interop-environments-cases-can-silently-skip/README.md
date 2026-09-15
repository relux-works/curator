# TASK-260906-2cfxfv: interop-environments-cases-can-silently-skip

## Description
Close the hole that lets a candidate root drop an environments vector family and still pass green. internal/interop reads five environments families behind guarded root-content skips and declares none of them in .github/ci/root-artifacts.tsv, so suite-plan.sh never defers the package, CI_REQUIRE_FULL_ROOT never fires for it, and root-content is policy allow in every lane including the candidate one. A root that stopped publishing vectors/environments.json would pass with all twenty-five environments cases silently skipped. Registering the package wholesale is not the answer: it would defer the pre-environments conformance cases the default lane depends on. Choose a shape that makes a missing family fatal in the candidate lane without losing default-lane coverage, and prove it with a negative run where the lane fails by name.

## Scope
curator repository, branch feat/interop-root-artifacts in worktree /Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage, base fb916acd60dd1881f050c5d838356235455b5b89. Authority curator-spec main 87a0d006. Files: .github/ci/root-artifacts.tsv, .github/ci/platform-cases.tsv, internal/interop and any package split it needs.

## Acceptance Criteria
A candidate root missing any environments vector family fails the candidate lane by name rather than skipping, proven by a run against a root with one family removed. The default lane keeps every non-environments internal/interop case it runs today. gate-selftest.sh green.
