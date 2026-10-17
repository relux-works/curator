# Producer brief: TASK-261017-0t5519, why the Windows test job takes 52 minutes

Read-only analysis. No product, test or workflow changes in this task.

The hosted `Test (windows-latest)` job takes about 52 minutes and is the critical path of every curator CI run (macOS Test about 16, macOS Race about 35, Linux Test about 2 on the self-hosted runners). Find where the time goes.

**Do:**
1. Take at least three recent green CI runs on main (`gh run list --branch main`), and for each the Windows, macOS and Linux test jobs. Use the job logs and any uploaded test evidence artifacts (the test gate writes Go test JSON under `.temp/ci-evidence/test`; check what the workflow uploads). Compute per-package and per-test wall time.
2. Name the top time sinks and their causes with evidence: filesystem and process-start cost, antivirus scanning of freshly built test binaries, real waits or sleeps and timeouts, packages that run serially, toolchain or cache misses, tests that do much more work on Windows.
3. Rank options to cut the time, each with expected gain and risk: targeted test changes, `-p`/`-parallel` or sharding the job, caching, a smaller subset on pull requests with the full suite on main, and a faster or self-hosted Windows runner. Do not propose dropping Windows coverage that guards a supported Windows feature without saying which feature loses coverage.
4. Recommend a runner size (vCPU, RAM, disk) if a self-hosted Windows machine is built.

**Output:** attach `TASK-261017-0t5519_windows-ci-analysis.md` with the tables, the evidence (run ids, job ids), the ranked options and the sizing recommendation. Public repository rules: no host names, people or local paths in board resources.

**Handoff:** there is no repository change; hand off with the analysis as the outcome and say so explicitly (empty repository delta is expected for this research task). Run `task-board handoff TASK-261017-0t5519 --role developer`; end your turn.
