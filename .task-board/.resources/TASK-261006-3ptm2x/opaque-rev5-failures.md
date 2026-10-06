# rev5 hosted gate 37493992848: failures (orchestrator extract). Ubuntu only; macOS and Windows are green.
Three NEW tests from rework 3 fail at their fixture setup, not at the NUL guard:
- internal/envprofile TestResolveAdmitsV2CommitPinnedContextNUL/v2-excluded and /v2-module (nul_opaque_v1_commit_test.go:281)
- internal/envprofile TestResolveCleanV1CommitContextStaysCurrent (nul_opaque_v1_commit_test.go:234)
Message: `environment_home_stale: passthrough entry .credentials.json is detached: link targets <tmp>/002/.credentials.json, expected <tmp>/...`
Reading: the fixture's operator HOME / credential link setup differs on Linux. Bare resolve sees the Claude `.credentials.json` passthrough link pointing at a different operator path from the one the profile expects. Compare with how passing envprofile tests pin the operator home (e.g. `pinOperatorHome(t, dir)` and the managedFixture helpers), and reuse that helper instead of a hand-rolled HOME. Assert the NUL behaviour after a clean resolve.
