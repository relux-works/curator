# BUG-261015-1tek9u: macos-race-flakes

## Description
Race (macos-latest) fails intermittently on tests unrelated to the change under test. Three distinct tests in two days: (1) cmd/curator TestCompiledProjectRepairsCorruptCompiledState/install; (2) buildrepo TestResolvedTransportTotalDeadlineBoundsSlowFetch, 15.8s over its deadline; (3) internal/install TestGlobalLaneRuntimeMatchesPackageMarker, git rev-parse refs/tags/v1^{commit} failing transiently on a fresh fixture repo. All passed on rerun. Suspected shared cause: wall-clock deadlines and git subprocess timing under -race on slow macOS runners.

## Scope
(define bug scope / affected area)

## Acceptance Criteria
Root cause found per test or a shared one; each test made deterministic (no wall-clock deadline under -race, fixture git state awaited) or its timing bound justified; 20 consecutive Race (macos-latest) runs green on main.
