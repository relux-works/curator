# TASK-261017-0t5519: windows-test-duration-analysis

## Description
The hosted Test (windows-latest) job takes about 52 minutes and is the critical path of every curator CI run (Linux Test on the self-hosted runners takes about 2 minutes; macOS Test 16, macOS Race 35). Find where the time goes and what would cut it.

## Scope
read-only analysis of hosted CI runs and the test suite; no product or workflow change in this task

## Acceptance Criteria
1. Per-package and per-test wall time of the Windows Test job from at least three recent green runs (go test JSON or the uploaded evidence), compared with macOS and Linux. 2. The top time sinks named with their cause (filesystem, process start, antivirus scanning, real waits or sleeps, serial packages, cache misses), each with evidence. 3. Ranked options to cut the time (test changes, sharding, parallelism, caching, PR versus main subsets, a faster runner), each with expected gain and risk. 4. A recommendation for the runner size if a self-hosted Windows machine is built.
