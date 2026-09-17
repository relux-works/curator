# TASK-260910-3ungjy — hosted gate failure on Change Request revision 1 (run 35220838841)

Extracted by the orchestrator from the CI evidence (`test-evidence-<os>`,
`test/go-test.json`). Two independent causes:

## 1. `cmd/curator TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands` — all lanes

The legacy-shape pin asserts `status --json` emits exactly the historical
three keys (`alias`, `path`, `skills`) for a closure without compiled
commands; the new `shell_hook_trust` array (two `approved`/`manager` rows
recorded by install) makes it four keys.

Orchestrator decision: the posture rows are spec-mandated (manager.md §8.6)
and an additive top-level key is the intended JSON extension. Update the pin
deliberately, not by loosening it: it must still prove that no `builds` key
appears (`object["builds"] == nil`) and now that the document is exactly
`alias`, `path`, `skills`, `shell_hook_trust` (count 4), with the rows'
closed shape (`path`, `state`, `approved_by` for recorded files) asserted.
Name the JSON addition in the CHANGELOG S6 entry ("`status --json` gains
`shell_hook_trust`"). Do not make the key conditional on warnings — an
all-approved posture is still posture.

## 2. `internal/hookapproval TestScanRejectsUnreadableState` and `TestAssessFailClosedOnUnreadableStateReportsUnapproved` — windows-latest only

Both new tests make the approval state "unreadable" by writing a regular
FILE at the manager-home path (`os.WriteFile(home, "not a directory")`), so
the state path `<home>/<state>` has a non-directory ancestor. On POSIX that
open fails with ENOTDIR (not `IsNotExist`), on Windows the same open fails
with ERROR_PATH_NOT_FOUND, which Go reports as `IsNotExist` — so the code
correctly (per §8.4) treats it as absence and the tests' premise is false on
Windows ("Scan over an unreadable state succeeded", "warnings = [], want the
unreadable state").

Make unreadability real on every platform instead of skipping: keep `home` a
directory and put a DIRECTORY at the state file's own path (reading a
directory as a file fails with a non-`IsNotExist` error on POSIX — EISDIR —
and on Windows — "The handle is invalid"/access error), or hold an exclusive
lock on Windows; keep the assertion that the failure is reported as the
unreadable state, never as absence. A GOOS-only skip is not acceptable; a
host-capability skip needs a named absent capability, and none is absent
here.

Re-run `go test -count=1 ./cmd/curator/... ./internal/hookapproval/...` and
hand off again; the runtime re-runs the hosted gate.
