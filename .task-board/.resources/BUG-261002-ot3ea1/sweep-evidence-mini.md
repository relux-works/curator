# Additional evidence: the Mac mini (macmini-infra, 2026-10-02)

Live processes on the Mac mini whose executable file was deleted by the cache sweeps of earlier `curator global upgrade` runs:
- tb-sessiond go-v1/59b546aa (board ~/.task-board): binary missing, process still running.
- tb-sessiond go-v1/3b57ed0a (board skill-project-management): binary missing, process still running.
- task-board go-v1/86ad9863: one process, binary missing.
- 6 task-board processes on the current 05ae996f: fine.

The bug affects both hosts. Long-lived daemons count too, so the sweep must retain ANY build that a live process executes, not only builds of daemons it knows about. Until the fix ships, the workaround is to relaunch through the plain shims.
