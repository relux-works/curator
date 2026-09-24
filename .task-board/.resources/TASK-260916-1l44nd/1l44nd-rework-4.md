# TASK-260916-1l44nd rework 4 (orchestrator, binding)

Revision 4 gate FAILED (run 35680994406) on the two ubuntu lanes with exactly TWO rows; Lint and
the other lanes are green:

1. `TestRunSessionTerminatesDescendants`: "the interpreter descendant did not start: open
   /dev/null: permission denied" — the confined interpreter (Landlock write-confinement applied)
   spawns a descendant with stdio redirected to `/dev/null`, and opening `/dev/null` for writing is
   now denied because no rule covers it. `/dev/null` (and the other standard device nodes the
   interpreter legitimately needs) must be reachable under confinement: add a file-typed rule
   (`READ_FILE|WRITE_FILE`) for `/dev/null` in the ruleset the worker installs — this is a property
   of the control, not of the row (any script that redirects to `/dev/null` would break) — and
   document it in the control's applied-rule set; keep everything else denied.
2. `TestLinuxMissingDerivedPathRefusesWriteConfinement` (`preflight_test.go:374`): the refusal is
   `script_execution_worker_protocol_invalid: cannot install inventory control
   "filesystem-write-confinement"` but the row expects it to NAME THE MISSING MEMBER (the derived
   path that does not exist). Make the install error carry the offending path (sanitized like the
   other diagnostics) and keep the class; adjust nothing in the row.
Continue from the revision-4 tree (no checkout/clean/stash); append "Revision 5" to results.md;
republish only on a green gate. No other change.
