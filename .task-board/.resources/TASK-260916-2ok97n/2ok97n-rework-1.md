# TASK-260916-2ok97n (R5) rework 1 → revision 2 — Windows real-interpreter defects (orchestrator brief, binding)

TOP PRIORITY: this is the final leaf of STORY-260822-2h0v9j, and its acceptance closes EPIC-260822. Revision 1's gate
(run 35861781844) was green on every lane EXCEPT `Test (windows-latest)`, and the Windows failures are REAL — exactly the
platform evidence R5 exists to produce. Fix them; change nothing else.

## Failure 1 — the ledger parser is not CRLF-safe
`internal/scriptpolicy` `TestScriptHostExecutionPolicySectionsAreAllClassified` and
`TestScriptHostExecutionPolicyProductionConsumersCoverAllCases` (conformance_test.go:264 and :327):
`platform-case ledger line 43 has 1 columns, want at least 6`. Line 43 of `.github/ci/platform-cases.tsv` is an EMPTY
line; the Windows runner checks the file out with CRLF, so the line reads `"\r"` and your parser (unlike the shell
gates) does not strip it. Fix the parser: strip a trailing `\r` and skip blank/comment lines exactly as the shell gate
does; add a unit row that feeds a CRLF ledger (with a blank line and a comment) and passes on every OS. Do NOT change
the ledger file to work around it.

## Failure 2 — declared `exec` of `cmd.exe` is not resolved on Windows
`internal/scriptworker` `TestWindowsRealInterpretersRunDeclaredExec` (windows_real_interpreter_test.go:81/87), with the
production search list (`ExecSearchDirs = nil` → `DefaultExecSearchDirs()`):
- python3-v1: `real python.exe run did not record the manager-resolved cmd.exe` → `Report.ResolvedExec["cmd.exe"]` empty;
- node-v1: `real node.exe exit code = 1 … Error: spawnSync cmd.exe ENOENT` → the interpreter's child cannot find cmd.exe.
Find the root cause with evidence (add temporary diagnostics to a disposable run if needed, then remove them):
where does resolution happen (manager vs worker process), which environment is visible there (is `SystemRoot`/
`SYSTEMROOT` present in that process — see `capabilities.go:383-430` and `exec.go:94-110`), and how the resolved path
reaches the script's PATH farm. Fix production minimally so a declared `exec` of `cmd.exe` resolves to the absolute
`%SystemRoot%\System32\cmd.exe`, is recorded in `ResolvedExec`, and is executable by the script through the manager's
PATH farm — without widening the environment beyond manager.md §3.1 (reserved names; SYSTEMROOT is manager-set on
Windows). Add a narrowing mutant (drop the Windows search dir or the resolution record) killed by this row.

## Rules
Keep every other lane green (the revision-1 gate was green on ubuntu/macos/race/lint/self-tests/interop). No test
weakened; the real-interpreter rows stay real (no fake `cmd.exe`). If python/node are genuinely absent on the runner the
existing skip reason applies — they were present in this run. Continue from the revision-1 tree (no
checkout/clean/stash), append "Revision 2" to `TASK-260916-2ok97n_results.md`, then
`task-board handoff TASK-260916-2ok97n --role developer`.
