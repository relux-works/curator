# TASK-260916-3dmjbc rework 2 (orchestrator, binding)

Verdict rev2: CHANGES_REQUESTED with ONE blocking finding
(TASK-260916-3dmjbc_review-verdict-rev2.md, RUN-260921-bd4d65). Continue from the revision-2
tree in the Story workspace (no checkout/clean/stash); fix exactly F1, then republish (rev 3)
on a green gate.

F1 (must) — `internal/scriptworker/interpreter.go:95-143` `readInterpreterIdentity` admits any
regular hashed file: a POSIX `#!/bin/sh` wrapper (pyenv/asdf/volta shims) passes both worker
verifications and the executed program becomes `/bin/sh` + whatever the wrapper execs (graph
nodes never identity-verified); on Windows a `.cmd`/`.bat` binding runs through `cmd.exe`
(Go executes batch files via cmd.exe and re-parses args), and an extensionless absolute path
is rewritten by Go's `lookExtensions` to `<path>.exe/.com/.bat/.cmd`, so the hashed file and
the executed file differ. Fix shape (reviewer-verified): require a native image header
through the SHARED godriver primitive (export the `nativeExecutableHeader`/`validateLauncher`
-style check from `internal/godriver` and call it from `readInterpreterIdentity`, which both
the manager-side `ResolveInterpreter` and the worker's pre-exec verification share — do not
fork it), and on Windows require the `.exe` extension so `lookExtensions` returns the path
unchanged. Rows: POSIX shebang wrapper binding → `script_execution_worker_identity_invalid`
at `ResolveInterpreter` AND at the worker (`expectFailure`, marker never written); Windows
`.cmd`/`.bat` and extensionless bindings refused (register on windows-latest like the existing
rows); mutant: drop the header check → both rows fail. Keep the diagnostics table and
docs/troubleshooting in step ("not a native interpreter image").

Also fold in cheaply if they touch the same lines (otherwise leave to R2/R3, recorded as
bounds in results.md): R-H (`scriptpolicy.MandatoryControls` exported mutable slice → copying
accessor). Do NOT start R-A (launcher), R-B (stream model), R-C (permit frame), R-F (runtime
area application) here — they are R2/R3 scope and will be briefed separately; restate them as
bounds in results.md "Revision 3" exactly as the verdict lists them so R2/R3 briefs can cite
them. Rerun the process-boundary rows on your host after the fix; append "Revision 3" to
results.md with the fix, rows, mutant and the updated implemented-controls table.
