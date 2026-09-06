# TASK-260906-2cfxfv: interop-environments-cases-can-silently-skip

## Description


## Scope
Stage (b) review cycle 3 (m5) confirmed and widened this: internal/interop declares none of its environments vector families in .github/ci/root-artifacts.tsv, so suite-plan.sh never defers the package and CI_REQUIRE_FULL_ROOT=1 never fires for it. A candidate root that dropped vectors/environments.json would pass the candidate lane green with all 25 environments cases silently skipped under the root-content class, which skip-classes.tsv marks policy allow in every lane. The reviewer names five further families covered by the same hole and traces it to stage (a) commit 4b5cd059. Stage (b) fixed the equivalent hole for internal/envfragment and internal/envmarker by registering their schema-case families, which the cycle-3 reviewer verified fails closed: suite-plan exits 1 in the candidate lane when either family is missing. internal/interop cannot simply be registered wholesale, because deferring the whole package would drop its unrelated coverage from the default lane; the environments conformance tests likely need their own package.

## Acceptance Criteria
A candidate root missing any environments vector family fails the candidate lane by name rather than skipping, proven by a run against a root with one family removed. The default lane keeps every non-environments internal/interop case it runs today. gate-selftest.sh green.
