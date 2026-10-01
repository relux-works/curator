# Review note — TASK-260930-3b3oyi rev3 re-review (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Re-review the newest CR (gate green) against your rev2 verdict and `3b3oyi-rework-1.md`. Verify:
1. F1: the posture marker is honoured only when its value equals os.Getppid().
   - Probe with real exit codes: `CURATOR_INTERNAL_POSTURE_WARNED=1 curator env status` and a wrong-pid value both still print the
     warning; a nested `curator run …` prints exactly one.
   - Rows (a)/(b)/(c) exist.
   - Kill the drop-Getppid mutant yourself.
   - The grandchild case is commented.
2. F2: with a config present, `curator env resolve --help` and `global add --help` behave exactly as on base 5ed5c4e1 (FlagSet usage,
   exit 2). Group help works without a config. The rows would catch the regression.
3. Items 1–4 from rev2 are still OK; no CHANGELOG/LOGBOOK; no stray files.
accept_cr, or changes requested with file:line. Never spell any employer name.
