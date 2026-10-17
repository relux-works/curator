# BUG-261017-ygfdal: landlock-probe-recursive-test-reexec

## Description
On Linux hosts with Landlock, probeScriptLandlock re-executes os.Executable() with the hidden mode __curator-script-landlock-probe. Inside a test binary that has no TestMain dispatch for that mode (observed: internal/install, install.test), the child ignores the mode and runs the whole test suite again, which probes again: a recursive process tree. The 30 s CommandContext timeout kills only the direct child, and the probe child gets the sanitized script-worker environment, so grandchildren escape the CI runner's job cleanup and keep running after the job ends. Observed 2026-10-17 on a self-hosted Linux runner: thousands of orphaned install.test probe processes, load average above 200 on 16 threads, after a cancelled Test (ubuntu-latest) job. Hosted runners hide it because their VM is discarded after the job.

## Scope
internal/scriptworker probe and every test package that can reach it; no workflow or host changes

## Acceptance Criteria
1. Re-executing any affected test binary in the hidden probe mode exits promptly with the probe result and runs no tests. 2. A probe child cannot start another probe. 3. Probe timeout or cancellation leaves no descendant processes on Linux. 4. Probe present/absent results are unchanged. 5. Regression test red on main, green with the fix; full remote gate green.
