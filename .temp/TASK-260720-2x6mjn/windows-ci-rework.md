# Windows CI rework for PR #15

## Context

- Repository: `git@github.com:ivanopcode/cocoaskills.git`
- Accepted task branch: `task/TASK-260720-2x6mjn-pure-build-planner`
- Accepted rebased commit: `323ea47db821641bc68f2b9054b50e82642a6df2`
- PR: https://github.com/ivanopcode/cocoaskills/pull/15
- Failed run: https://github.com/ivanopcode/cocoaskills/actions/runs/30551952750
- macOS, Ubuntu, and strict mypy passed. All four Windows Python jobs failed with the same seven tests.

## Windows failures

1. `tests/test_build_planner.py::test_filesystem_generation_probe_is_deterministic_and_read_only`
   - `BuildPlanningError: concurrent_state_change: shared planning file changed while opening`
   - The generation/fingerprint probe is not stable under Windows file metadata semantics.
2. `tests/test_global_install.py::test_global_install_dry_run_does_not_write_anywhere`
3. `tests/test_global_install.py::test_global_dry_run_missing_go_fails_whole_plan_without_mutation`
4. `tests/test_global_install.py::test_global_build_planning_rejects_conflicts_across_isolated_closures`
5. `tests/test_global_install.py::test_global_upgrade_dry_run_does_not_create_or_fetch_skills_root`
   - These report the same false `concurrent_state_change` for `Skillfile.json`.
6. `tests/test_global_install.py::test_global_real_install_does_not_plan_builds_or_require_go`
   - Fixture replaces `PATH`; on Windows its POSIX symlink population finds no `git`, so real install reports `git executable not found`.
7. `tests/test_install.py::test_real_install_does_not_plan_builds_or_require_go`
   - Same fixture behavior produces `[WinError 2]`.

## Required rework

- Preserve all accepted planner behavior and the stable error contract.
- Make the read-only filesystem generation probe deterministic on Windows without weakening concurrent-change detection.
- Make the two no-Go real-install tests portable: preserve the minimum trusted executable set needed by existing install behavior while still proving Go is absent.
- Add focused regression coverage for the Windows semantics where practical.
- Do not modify workflow matrices or skip/xfail these tests.
- Run focused tests and strict mypy locally. Push the updated task branch only to `origin` (`ivanopcode/cocoaskills`), never `wb`.
- Leave the task at `to-review` with changed-file and test evidence. Do not merge or land.
