# TASK-260824-1kzj22: parallelize-cmd-curator-test-suite

## Description
Wall-clock of the full repository gate is dominated by cmd/curator: 11 test files, 70 Test functions, zero t.Parallel(), ~10 minutes sequential. Mark independent tests t.Parallel() with proper isolation so producer/reviewer full-suite gates stop paying the sequential queue.

## Scope
Add t.Parallel() to independent tests in cmd/curator with per-test isolation (t.TempDir, no shared env/cwd/global state, no port or fixture collisions). Keep godriver timeout behavior intact (-timeout 30m stays). Do not change production code. Splitting into sub-packages or -run sharding is out of scope unless t.Parallel() alone cannot reach the target.

## Acceptance Criteria
Full cmd/curator package wall-clock reduced to at most 4 minutes on the reference 8-core machine with unchanged test coverage and zero new flaky failures across three consecutive uncached runs (go test -timeout 30m -count=1 ./cmd/curator), race run stays green (go test -race focused subset), and evidence logs attached.
