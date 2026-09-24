# TASK-260916-1l44nd rework 3 (orchestrator, binding)

Revision 3 gate FAILED (run 35676751063) — enforcement now works; three residual failures on
`Race (ubuntu-latest)` (Test ubuntu green) plus lint:

1. `TestOperationPrivateRuntimeAreaApplied/paths-derive-project-root` (Race ubuntu):
   `capability_evidence_invalid: host-conditional control "descendant-exec-denial" reports status
   "unavailable" against this invocation's probe` — the probe said the control is present but the
   worker's evidence says unavailable for THIS invocation (or vice versa) only in the row that
   derives a filesystem path set. Under `-race` timing differs, so this is a real
   probe↔apply divergence: the exec-denial decision must be made ONCE per invocation from the same
   probe result the record cites (no re-probe in the worker, no per-path fallback that silently
   downgrades the control when a rule cannot be added — if a derived path cannot be ruled, that is
   an apply failure of the control reported consistently, or an invocation refusal, never a
   status flip). Make the row deterministic under `-race`.
2. `TestRunSessionTerminatesDescendants` (Race ubuntu): "the interpreter descendant did not
   publish its identity: open …/descendant.pid: no such file" — with Landlock write-confinement
   applied, the fixture descendant can no longer write its pid file into the test temp dir
   (outside the derived path set), so the row waits 14 s and fails. Fix the ROW: publish the
   descendant identity through an allowed channel (the stdout/protocol result, the runtime area,
   or a path inside the derived set), or declare the pid path in the row's `filesystem`
   declaration. Do not weaken confinement.
3. Lint: `landlock_linux.go:212,215` ST1005 error strings must not be capitalized.
Audit every other row that writes side files from inside the confined interpreter for the same
assumption (Linux lanes). Continue from the revision-3 tree (no checkout/clean/stash); append
"Revision 4" to results.md; republish only on a green gate.
