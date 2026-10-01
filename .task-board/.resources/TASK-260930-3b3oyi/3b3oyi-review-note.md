# Review note — TASK-260930-3b3oyi second-operator first-run fixes (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Review the newest CR (gate green) against `secondfix-brief.md` and `3b3oyi-gatefix-1.md`. These are findings 1, 2, 4, 5 and 6 of the
second-operator walk (TASK-260930-o5uu7a_results.md). Verify each at the production entry, with the exact messages and real exit
codes, on a throwaway HOME with binaries built from the candidate:
1. On a clean HOME, `curator global init` still refuses, and the message names `curator bootstrap --skills-root <dir>`. `--help` works
   without global config. docs/cli.md states that bootstrap is the prerequisite.
2. The docs/cli.md bootstrap example works as written.
3. `env status --check`: only current-scope rows decide the exit code. Other profiles' findings are still printed.
4. `env resolve` without --repair on a fresh profile: the diagnostic says "rerun with --repair", and the exit code is unchanged.
5. The posture warning appears once per process tree, including under `curator run`, and still appears once. Check the internal
   marker cannot be abused to hide the warning for a user-started command, i.e. it is only honoured when set by curator itself.
6. Check the Windows fix from gatefix-1. No weakened assertions; no CHANGELOG/LOGBOOK; no stray files.
accept_cr, or changes requested with file:line. Never spell any employer name.
