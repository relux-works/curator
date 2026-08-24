# TASK-260824-1kzj22: parallelize-cmd-curator-test-suite

## Description
Wall-clock of the full repository gate is dominated by cmd/curator: 11 test files, 70 Test functions, zero t.Parallel(), ~10 minutes sequential. Mark independent tests t.Parallel() with proper isolation so producer/reviewer full-suite gates stop paying the sequential queue.

## Scope
Add t.Parallel() to independent tests in cmd/curator with per-test isolation (t.TempDir, no shared env/cwd/global state, no port or fixture collisions). Keep godriver timeout behavior intact (-timeout 30m stays). Do not change production code. Splitting into sub-packages or -run sharding is out of scope unless t.Parallel() alone cannot reach the target.

## Acceptance Criteria
AMENDED after measurement (reviewer RUN-260824-69b8fd): the original <=4 min wall-clock target is below the measured single-test floor and is not a meaningful target for a test-only task. DELIVERED: t.Parallel() on 36 of 69 independent cmd/curator cases (reviewer-verified independent via transitive call-closure scan, zero process-global reachability, coverage unchanged at 57.1%, no new flakes across 3 uncached runs, focused -race green, lint/gofmt/vet clean). Measured floor ~373s, dominated by TestCompiledProjectStatusRepairRollbackRecovery at 220-270s (five subtests sharing one installed fixture; cost is godriver cold compilation with hermetic per-session GOCACHE, not test overhead). Getting genuinely under 4 min requires a production refactor (injectable config + capture) out of this task-only scope, tracked as a follow-up. The closest-reachable-with-bottleneck-named branch of the original AC IS met.
