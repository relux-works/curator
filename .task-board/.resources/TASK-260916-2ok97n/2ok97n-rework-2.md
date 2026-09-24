# TASK-260916-2ok97n (R5) — rework 2 (orchestrator, binding). HIGHEST PRIORITY in the campaign.

> **THIS IS THE ONLY CURRENT INSTRUCTION.** Every other brief attached to this task (2ok97n-brief, rework-1, handoff-2, reconciliation) is history/context. A previous run followed the old handoff brief and changed no code — that is not acceptable: this run MUST change production code as below before any handoff.

Revision 2's hosted gate (run 35885026680): everything green except Windows. The CRLF ledger defect is FIXED
(no ledger failure). Two Windows failures remain:
1. `internal/snapshot TestConcurrentGetAcceptsOneImmutablePublisher` — known flake, fixed by BUG-260923-11jgkt
   which is landing now. NOT yours; do not touch internal/snapshot.
2. YOURS — `internal/scriptworker TestWindowsRealInterpretersRunDeclaredExec` (evidence from the gate's
   `test-evidence-windows-latest/test/go-test.json`):

       manager SYSTEMROOT="C:\\Windows"; exec search dirs=["C:\\Windows\\System32" "C:\\Windows"]; declared cmd.exe target="C:\\Windows\\System32\\cmd.exe"
       python3-v1: real python.exe run did not record the manager-resolved cmd.exe; ... farm entries=["python.exe"]
       node-v1: spawnSync cmd.exe ENOENT (PATH farm has no cmd.exe)

   So the pure search list is right, but on the real `Launch` path with `request.ExecSearchDirs = nil`,
   `Report.ResolvedExec` is EMPTY and the farm holds only the interpreter. Python "passed" stdout only because
   CreateProcess falls back to System32 without PATH — that is ambient resolution, not the manager's.

## Do
a. Trace, in code, `Launch(request)` → capability derivation → declared-exec resolution → farm build, for a
   request whose `ExecSearchDirs` is nil and whose declared capabilities are `{"exec":["cmd.exe"]}` exactly as
   the test's `fixture.declareCapabilities` writes them. Find where the declared exec is dropped (candidates:
   nil `ExecSearchDirs` meaning "none" instead of "default" on this path; the fixture's declaration not being
   the one the derivation reads; name matching `cmd.exe` vs PATHEXT stripping/appending; the Windows farm copy
   path skipping non-interpreter entries). State the root cause with file:line.
b. Fix it in production code (not the test). Keep the existing Darwin/Linux behaviour byte-identical.
c. FAIL CLOSED: a declared exec that the manager cannot resolve must refuse the launch before the worker
   starts (typed error), never start a worker whose farm lacks a declared exec. Add a cross-platform row for
   that refusal at the production entry (`Launch`), and a cross-platform row that drives the SAME code path the
   Windows test uses (nil `ExecSearchDirs` → default list) with an injected GOOS/env seam or the existing pure
   helpers, asserting `ResolvedExec` and `FarmEntries` are populated — so this is provable on macOS/Linux, not
   only on the Windows lane.
d. Narrowing mutants (disposable copy only): drop the declared exec from the farm; make nil ExecSearchDirs mean
   empty — each must be killed by a cross-platform row. Table in results.
e. Bounded local runs: `go test ./internal/scriptworker/...` (and `-race`), cross-compile the Windows test binary.

Append "Revision 3" to results (root cause file:line, fix, rows, mutants), then
`task-board handoff TASK-260916-2ok97n --role developer`. A `run_wrote_outside_worktree … policy warn` block is a
warning — verify the status moved to `to-review`. No LOGBOOK.md.
