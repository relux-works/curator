# TASK-260916-1h82gq rework 2 (orchestrator, binding)

Revision 2 gate FAILED only on `Test (windows-latest)` (run 35633381854) and only in
`TestCapabilityDerivationAllFieldsAbsentDenyByDefault` (both subtests): the interpreter's
environment keys contain `SYSTEMROOT` while the row expects the manager-built set without it.
Cause: Go's `os/exec` on Windows injects `SYSTEMROOT` into a child's environment when the
explicit `cmd.Env` lacks it (needed for the Windows loader; `os/exec` docs: "On Windows, ... adds
SYSTEMROOT if missing"). The spec reserves `SYSTEMROOT`/`WINDIR` as manager-owned names, so the
manager-built environment on Windows MUST set them itself (manager-set values, from the host,
never from `env_read`), and the row must expect them on Windows. Fix: the environment builder
adds the Windows process essentials the loader needs (`SYSTEMROOT`, and `WINDIR` if you set it
deliberately) as manager-set values on Windows only; the row's expected set is per-OS; document
the reserved-name handling in results.md ("manager-set on Windows: SYSTEMROOT …"). No other
change. Continue from the revision-2 tree (no checkout/clean/stash); append "Revision 3" to
results.md; republish only on a green gate.
