# TASK-260916-2ok97n (R5) — rework 3 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION. HIGHEST PRIORITY.

Revision 3 (refreshed onto fad88136) failed on ALL OSes in one test, and the fault is the orchestrator's rework-2 brief:

    internal/install TestEnforcedScriptCommandInstallsNativeLauncher: launcher exit = 1,
    "script-worker-v1 script_execution_declared_exec_unresolved: manager could not resolve declared exec name(s): no-such-tool-zzz-cur"

The spec is explicit (curator-spec protocol/core.md, section on derived capabilities, lines ~329-337): "A declared name the
manager cannot resolve is ABSENT from the built PATH and MUST be REPORTED; it MUST NOT be resolved from the caller's PATH
at launch or by the interpreter at run time." There is no launch refusal. So:
1. REMOVE the `script_execution_declared_exec_unresolved` launch refusal from `DeriveProfile`/`Launch`. Instead the
   unresolved name is absent from the farm and REPORTED on the existing report/audit surface (name the field; add one
   if none exists, e.g. `Report.UnresolvedExec`). Keep it impossible to reach the name through the caller's PATH.
2. Replace `TestLaunchRefusesUnresolvedDeclaredExecBeforeWorker` with a row asserting: launch proceeds, the name is
   absent from the farm, it is reported, and a same-name executable on the caller PATH is NOT reachable by the script.
   `TestEnforcedScriptCommandInstallsNativeLauncher` must pass unchanged.
3. KEEP the rest of revision 3: nil ExecSearchDirs → manager-derived default list; the Windows System32 identity fix.
   On the hard-link allowance: core.md says exec names resolve "under the interpreter resolution rules above", which
   reject "hard-link substitution". State in results exactly how your allowance is bounded (default Windows list only,
   canonical `%SystemRoot%\System32` from the manager's captured SYSTEMROOT, target physically below it) and why a
   component-store link there is not a substitution; the orchestrator files a spec erratum to make that explicit.
4. Mutants (disposable copy): report dropped for an unresolved name → killed; caller-PATH fallback for an unresolved
   name → killed; keep the three rev-3 mutants killed.
Bounded local runs: `go test ./internal/scriptworker/... ./internal/install -run 'Script|Enforced'` (+ `-race` on
scriptworker), gate scripts. Append "Revision 4" to results, then `task-board handoff TASK-260916-2ok97n --role
developer`; stay in the turn while the gate runs. A `run_wrote_outside_worktree … policy warn` block is a warning —
verify status `to-review`.
