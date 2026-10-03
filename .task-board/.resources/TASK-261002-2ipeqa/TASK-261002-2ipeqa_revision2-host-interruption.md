# TASK-261002-2ipeqa — Revision 2 host interruption

Worktree base: 64345d71. Revision 2 removed 12 stale rc.14 gaps (five marker cross-field cases in each of marker v3 and v4, plus two Windows executable identity cases), and four stale Windows rows in earlier hash-v2/Muse candidate digests. rc.14 retains nine posture/provisioning gaps. Count pins remain 103 families / 1,870 case entries. Writer stays OFF and SPEC_PIN remains rc.13. LOGBOOK.md was not edited.

Observed standalone validations under GOFLAGS=-work:
- Candidate required command: exit 0, 22.436 seconds, syspolicyd running before/after, successive crashes 366 -> 366.
- Default-pin required command: exit 0, 16.202 seconds, syspolicyd running before/after, successive crashes 366 -> 366.
- Corpus authentication audit: exit 0, 0.470 seconds, 103 families / 1,870 case entries / nine gap rows, crashes 366 -> 366.

The marker/scriptworker wrapper started with syspolicyd running and crash count 366. It stopped responding during its AFTER-test launchctl observation. After more than ten minutes of blocked diagnostics, it was interrupted with Ctrl-C; the wrapper's real exit code was 130. The underlying Go process had returned before the observation, but its numeric exit code was not persisted. It is UNKNOWN and is not claimed as a green validation. The raw test log remains marker-scriptworker.log. Do not infer an exit code from its output.

The independent 103-family count-audit wrapper never printed its preflight observation and was interrupted with exit 130. This attempt is not a pass. Both extra ps probes were interrupted with exit 130; two interactive shell probes were interrupted with exit 1. A further bash built-in echo probe is pending. New process launches, including shells with only built-ins, remain unresponsive. The host-rules five-minute backoff elapsed without recovery. No later Go gate was started.

The exact Implementations check, its presence/consumption gates, fresh lint, build, and independent all-family count audit remain unrun for Revision 2. Historical Revision 1 evidence remains historical; it is not accepted as validation of this changed tree.

Recovery needed: restore host process execution and launchctl responsiveness, then obtain a fresh syspolicyd state/crash count, observe the five-minute backoff if required, rerun the missing checks and attach this Revision 2 evidence through task-board before developer handoff. No product decision or workaround is needed.

Recovery observed at 2026-10-02 09:48 UTC: shells and launchctl respond; syspolicyd is running, successive crashes 367 (previously 366). The saved Go log shows marker PASS and scriptworker FAIL after its 10-minute timeout in TestRunShimToleratesUnusableDiagnosticsDir. Numeric Go exit remains unknown; wrapper exit 130 is retained. A five-minute backoff is being observed before further Go commands. The independent all-family audit subsequently exited 0 after two audit-script field-mapping failures (each exit 1). It authenticated and recounted 103/103 families / 1,870 entries.

Second recurrence: the shared build lock became available and the exact marker/scriptworker rerun started at about 10:00:30 UTC, with syspolicyd running and crash count 367. By 10:01 UTC, new diagnostic shell launches again stalled. No numeric Go exit was observed. Further Go attempts were stopped. This reproduces the external host blocker despite GOFLAGS=-work, exclusive build-lock ownership, and more than five minutes of backoff after the first detected crash. No product code or test workaround was introduced. The remaining required marker/scriptworker validation, exact Implementations rerun, and fresh lint/build cannot be attested on this host until its process-launch problem is resolved. Restore host execution or supply a working validation host; retain the same frozen source tree and candidate/default roots, rerun the missing commands, then attach evidence before developer handoff.

Second recovery observed at 2026-10-02 10:08 UTC: syspolicyd running, successive crashes 368 (367 before the exact retry). No later Go validation started. Both interruptions reproduced the host failure despite -work, exclusive lock ownership and required backoff. Task is blocked pending a working validation host.
