# TASK-260930-3b3oyi — second-operator first-run fixes (THE ONLY CURRENT INSTRUCTION; priority)

Source: TASK-260930-o5uu7a_results.md, findings 1, 2, 4, 5 and 6. The walk ran on a clean HOME. Fix these, each with a test at the
production entry:
1. `curator global init` (and `global --help`, `profile *`, `env *`) on a clean HOME exits 1 with `global config not found`.
   Keep the refusal. Make the message say how to fix it:
   "run `curator bootstrap --skills-root <dir>` first (see docs/cli.md#bootstrap)". `--help` must print help without requiring
   global config. Update docs/cli.md:603-616 to state that bootstrap is the prerequisite.
2. The docs/cli.md:44 example `curator bootstrap --if-missing --non-interactive` exits 2 (`bootstrap requires --skills-root`). Fix the
   example.
3. `env status --check` exits 1 on a healthy setup, because non-current profiles report `home unprovisioned`. Only current-scope rows
   decide the exit code. Other profiles' findings are still printed as informational.
4. `env resolve <env>` without --repair on a fresh profile fails with `environment_home_stale: home unprovisioned` and gives no hint.
   Append "rerun with --repair" to that diagnostic. The exit code is unchanged.
5. The `security_posture_permissive` warning prints on every command, and twice on nested invocations such as `curator run` → resolve.
   Print it at most once per process tree, e.g. suppress it in the nested resolve via an internal env marker. It must still print once.
Rows: the exact messages and exit codes for 1, 4 and 5; the --check exit code for 3. Also run `go test ./cmd/curator -run 'Bootstrap|GlobalInit|EnvStatus|EnvResolve|Posture'`
with real exit codes. No CHANGELOG/LOGBOOK edit: put the entry text in the results. Never spell any employer name. Update the results,
then run `task-board handoff TASK-260930-3b3oyi --role developer`, then END YOUR TURN.
