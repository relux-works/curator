# Review note for TASK-260916-3dmjbc revision 4 (orchestrator, binding)

Revision 4 = rework 2 + rework 3 for your revision-2 verdict (TASK-260916-3dmjbc_review-verdict-rev2.md;
briefs 3dmjbc-rework-2.md, 3dmjbc-rework-3.md). Expected delta since rev2: F1 only — interpreter
identity requires a native image header through the SHARED godriver primitive (exported, not
forked) and `.exe` on Windows; rows: POSIX shebang-wrapper binding refused at `ResolveInterpreter`
AND at the worker, Windows `.cmd`/`.bat`/extensionless bindings refused (registered on
windows-latest), mutant "drop the header check" kills both; diagnostics/troubleshooting in step;
optionally R-H (copying accessor for `MandatoryControls`); rev3 → rev4 only made the
`TestScriptWorkerRejectsTamperedInterpreter` fixture a real `.exe` on Windows (and audited the
other rows). Gate green on all lanes: run 35590341220 — verify the gate commit resolves to the
exact revision-4 tree.

Judge: (1) `git diff` rev2→rev4 is exactly that scope (list any other change and judge it);
(2) rerun your rev2 wrapper probe (`#!/bin/sh` wrapper bound as `python3-v1`) against rev4 —
refused before the interpreter runs, at both sites; (3) the godriver primitive is shared, not
copied, and the go-v1 rows are unchanged; (4) Windows rows present in the gate's Windows
evidence (extract the new row names from `test-evidence-windows-latest`); (5) the
implemented-controls table still truthful. Everything else was accepted at rev2 — do not
re-open it unless rev4 changed those bytes. Residuals R-A…R-J stay recorded for R2/R3/R5.
Record exactly one verdict: accept_cr(TASK-260916-3dmjbc, revision=4, evidence=<your outcome
resource>) on ACCEPT, or changes_requested with file:line and reproduction. Bounded local
commands; host stall windows — retry a hung command once.
