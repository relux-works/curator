# TASK-260916-3dmjbc rework 3 (orchestrator, binding)

Revision 3 gate FAILED only on `Test (windows-latest)` (run 35586991022), one row:
`internal/scriptworker :: TestScriptWorkerRejectsTamperedInterpreter` —
`worker_test.go:462: script-worker-v1 script_execution_worker_identity_invalid: the resolved
interpreter node-v1 binding must name the native .exe image itself`. Your F1 fix (Windows
bindings must name a `.exe` native image) now trips on this row's fixture BEFORE the tamper
it wants to prove: the interpreter fixture on Windows is not a `.exe`/native image, so the row
observes the wrong refusal reason (or observes a refusal where it expects the pre-tamper accept
+ post-tamper refusal). Fix the ROW, not the check: on Windows build the interpreter fixture as
a real `.exe` (the stub interpreter is already compiled per test binary — name it with the
platform suffix, as the manager copies are), let the pre-tamper resolution succeed, then tamper
the bytes and assert the identity refusal with the tamper-specific reason. Audit every
scriptworker row that binds an interpreter fixture for the same assumption on Windows so the
next gate does not fail on the next one (rev1 and rev3 each cost a gate for the same class).
Continue from the revision-3 tree (no checkout/clean/stash); append "Revision 4" to results.md
(the row change, Windows proof status); republish only when the gate is green. Rulings of
3dmjbc-brief.md and 3dmjbc-rework-2.md unchanged.
