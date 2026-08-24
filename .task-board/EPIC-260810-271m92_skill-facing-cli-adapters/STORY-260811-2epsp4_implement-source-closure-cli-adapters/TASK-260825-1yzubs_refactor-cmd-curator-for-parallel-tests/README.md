# TASK-260825-1yzubs: refactor-cmd-curator-for-parallel-tests

## Description
Follow-up from TASK-260824-1kzj22 review RUN-260824-69b8fd. The remaining 33 serial cmd/curator tests cannot be parallelized without production changes: run() reads process-global CURATOR_CONFIG (main.go:94 -> config.Load("")) and writes process-global os.Stdout/os.Stderr via capture. A production refactor to make config and output capture injectable per-invocation would let those tests use t.Parallel(). Additionally TestCompiledProjectStatusRepairRollbackRecovery alone is 220-270s (five subtests share one installed fixture; cost is godriver cold compilation with hermetic per-session GOCACHE) and needs splitting or build-cost reduction to get the package under 4 minutes.

## Scope
Make cmd/curator run() take an injectable config source and injectable stdout/stderr (instead of process-global CURATOR_CONFIG and os.Stdout/os.Stderr capture) so per-test state is hermetic; then parallelize the previously-serial tests. Separately split or speed up TestCompiledProjectStatusRepairRollbackRecovery. If codex/legacy-board-repair merges to main, also inventory and parallelize the 7 main-only cmd/curator cases the 1kzj22 review named (toolchain_remedy_test.go plus six main_test.go cases). Production code changes are IN scope here (unlike 1kzj22).

## Acceptance Criteria
cmd/curator full-package wall-clock under 4 minutes on the reference 8-core machine with unchanged coverage and zero new flaky failures across three consecutive uncached runs; config and output-capture seams are injectable and documented; the heavy compiled-status test no longer dominates; race run green; evidence attached.
